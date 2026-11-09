package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	backupmodel "github.com/cloud-barista/ai-ha/pkg/backup/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// These run against a real PostgreSQL, like the rest of the store's tests: what
// matters is that the filters reach the statement and that a job with no
// attempt survives being read alongside ones that have one.

// enqueueFor creates a job for a system of its own, so that a test can select
// exactly its rows from a shared database.
func enqueueFor(t *testing.T, svc *Service, pool *pgxpool.Pool, systemID string) Job {
	t.Helper()
	job, err := svc.CreateBackup(context.Background(), systemID,
		BackupSpec{Devices: []string{"/dev/sda1"}, RepositoryID: testRepoID}, "test")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM jobs WHERE job_id = $1`, job.JobID)
	})
	return job
}

// finish runs a job to the given outcome and reports detail.
func finish(t *testing.T, svc *Service, j Job, outcome Outcome, detail json.RawMessage) {
	t.Helper()
	ctx := context.Background()
	who := Identity{AgentID: testAgentID, SystemID: j.SystemID}
	claimed, err := svc.Claim(ctx, who, j.JobID, testClaimKey)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if _, err := svc.RecordProgress(ctx, who, j.JobID, claimed.Attempt.AttemptID, claimed.LeaseToken,
		Progress{Seq: 1, ProcessedBytes: 4096, Advanced: true}); err != nil {
		t.Fatalf("progress: %v", err)
	}
	if _, err := svc.RecordResult(ctx, who, j.JobID, claimed.Attempt.AttemptID, claimed.LeaseToken,
		Result{Outcome: outcome, Message: "done", Detail: detail}); err != nil {
		t.Fatalf("result: %v", err)
	}
}

// setDetail reports a set with a new random ID, so that repeated runs against
// the same database cannot find each other's sets.
func setDetail(t *testing.T) (string, json.RawMessage) {
	t.Helper()
	id := backupmodel.NewID()
	return id, json.RawMessage(fmt.Sprintf(`{"backupSetId":%q,"status":"completed","items":3}`, id))
}

func TestListFilters(t *testing.T) {
	svc, _, pool := testService(t)
	ctx := context.Background()
	system := "sys-" + backupmodel.NewID()
	queued := enqueueFor(t, svc, pool, system)
	done := enqueueFor(t, svc, pool, system)
	_, detail := setDetail(t)
	finish(t, svc, done, OutcomeCompleted, detail)
	other := enqueueFor(t, svc, pool, "sys-"+backupmodel.NewID())

	ids := func(f ListFilter) map[string]bool {
		t.Helper()
		list, err := svc.List(ctx, f)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		out := make(map[string]bool, len(list))
		for _, j := range list {
			out[j.JobID] = true
		}
		return out
	}

	got := ids(ListFilter{SystemID: system})
	if len(got) != 2 || !got[queued.JobID] || !got[done.JobID] || got[other.JobID] {
		t.Errorf("by system: %v", got)
	}
	if got := ids(ListFilter{SystemID: system, Status: StatusCompleted}); len(got) != 1 || !got[done.JobID] {
		t.Errorf("by status: %v", got)
	}
	if got := ids(ListFilter{SystemID: system, Kind: KindBackup}); len(got) != 2 {
		t.Errorf("by kind: %v", got)
	}
	if got := ids(ListFilter{SystemID: system, Kind: Kind("restore")}); len(got) != 0 {
		t.Errorf("by another kind: %v", got)
	}

	future := time.Now().Add(time.Hour)
	if got := ids(ListFilter{SystemID: system, From: future}); len(got) != 0 {
		t.Errorf("from the future: %v", got)
	}
	if got := ids(ListFilter{SystemID: system, To: future}); len(got) != 2 {
		t.Errorf("up to the future: %v", got)
	}
}

// Walking the pages must visit every job once, which needs an order that does
// not depend on two jobs having different creation times.
func TestListPagesVisitEveryJobOnce(t *testing.T) {
	svc, _, pool := testService(t)
	ctx := context.Background()
	system := "sys-" + backupmodel.NewID()
	const total = 5
	for range total {
		enqueueFor(t, svc, pool, system)
	}
	// The same instant, as a bulk insert would give them.
	if _, err := pool.Exec(ctx, `UPDATE jobs SET created_at = now() WHERE system_id = $1`, system); err != nil {
		t.Fatalf("align times: %v", err)
	}

	seen := make(map[string]bool)
	for offset := 0; offset < total; offset += 2 {
		page, err := svc.List(ctx, ListFilter{SystemID: system, Limit: 2, Offset: offset})
		if err != nil {
			t.Fatalf("page at %d: %v", offset, err)
		}
		for _, j := range page {
			if seen[j.JobID] {
				t.Fatalf("job %s came back twice", j.JobID)
			}
			seen[j.JobID] = true
		}
	}
	if len(seen) != total {
		t.Fatalf("saw %d jobs, want %d", len(seen), total)
	}
}

func TestListBackupsCarriesAttempts(t *testing.T) {
	svc, _, pool := testService(t)
	ctx := context.Background()
	system := "sys-" + backupmodel.NewID()
	waiting := enqueueFor(t, svc, pool, system)
	done := enqueueFor(t, svc, pool, system)
	setID, detail := setDetail(t)
	finish(t, svc, done, OutcomeCompleted, detail)

	records, err := svc.ListBackups(ctx, ListFilter{SystemID: system})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("%d records, want 2", len(records))
	}
	for _, r := range records {
		switch r.Job.JobID {
		case waiting.JobID:
			if r.Attempt != nil {
				t.Errorf("a job nobody claimed has an attempt: %+v", r.Attempt)
			}
		case done.JobID:
			if r.Attempt == nil {
				t.Fatal("a finished job came back without its attempt")
			}
			if r.Attempt.ProcessedBytes != 4096 {
				t.Errorf("processed = %d", r.Attempt.ProcessedBytes)
			}
			var got struct {
				BackupSetID string `json:"backupSetId"`
			}
			if err := json.Unmarshal(r.Attempt.Result, &got); err != nil || got.BackupSetID != setID {
				t.Errorf("result = %s (%v)", r.Attempt.Result, err)
			}
		default:
			t.Errorf("unexpected job %s", r.Job.JobID)
		}
	}
}

func TestGetBackupFindsASetThatFailed(t *testing.T) {
	svc, _, pool := testService(t)
	ctx := context.Background()
	failed := enqueueFor(t, svc, pool, "sys-"+backupmodel.NewID())
	setID, detail := setDetail(t)
	finish(t, svc, failed, OutcomeFailed, detail)

	got, err := svc.GetBackup(ctx, setID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Job.JobID != failed.JobID || got.Job.Status != StatusFailed || got.Attempt == nil {
		t.Fatalf("got %+v", got)
	}

	// It is on record, and still not something a restore may use.
	if _, err := svc.FindBackupSet(ctx, setID); !errors.Is(err, ErrBackupSetNotFound) {
		t.Fatalf("FindBackupSet: err = %v, want ErrBackupSetNotFound", err)
	}
}

func TestGetBackupUnknown(t *testing.T) {
	svc, _, _ := testService(t)
	if _, err := svc.GetBackup(context.Background(), backupmodel.NewID()); !errors.Is(err, ErrBackupSetNotFound) {
		t.Fatalf("err = %v, want ErrBackupSetNotFound", err)
	}
}

func TestGetBackupRefusesAnAmbiguousSet(t *testing.T) {
	svc, _, pool := testService(t)
	setID, detail := setDetail(t)
	finish(t, svc, enqueueFor(t, svc, pool, "sys-"+backupmodel.NewID()), OutcomeCompleted, detail)
	finish(t, svc, enqueueFor(t, svc, pool, "sys-"+backupmodel.NewID()), OutcomeCompleted, detail)

	if _, err := svc.GetBackup(context.Background(), setID); !errors.Is(err, ErrBackupSetAmbiguous) {
		t.Fatalf("err = %v, want ErrBackupSetAmbiguous", err)
	}
}

// A job that finished must not read as one somebody cancelled.
func TestCancelFinishedLeavesTheJobAlone(t *testing.T) {
	svc, _, pool := testService(t)
	ctx := context.Background()
	done := enqueueFor(t, svc, pool, "sys-"+backupmodel.NewID())
	_, detail := setDetail(t)
	finish(t, svc, done, OutcomeCompleted, detail)

	got, err := svc.RequestCancel(ctx, done.JobID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if got.Status != StatusCompleted || got.CancelRequested {
		t.Fatalf("status = %s, cancelRequested = %t", got.Status, got.CancelRequested)
	}
}

// Asking twice must not clear the request the first one made.
func TestCancelRepeatedKeepsTheRequest(t *testing.T) {
	svc, _, pool := testService(t)
	ctx := context.Background()
	j := enqueueFor(t, svc, pool, testSystemID)
	if _, err := svc.Claim(ctx, identity(), j.JobID, testClaimKey); err != nil {
		t.Fatalf("claim: %v", err)
	}

	for i := range 2 {
		got, err := svc.RequestCancel(ctx, j.JobID)
		if err != nil {
			t.Fatalf("cancel %d: %v", i, err)
		}
		if got.Status != StatusCancelling || !got.CancelRequested {
			t.Fatalf("cancel %d: status = %s, cancelRequested = %t", i, got.Status, got.CancelRequested)
		}
	}
}
