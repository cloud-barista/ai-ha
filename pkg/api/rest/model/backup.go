package model

import "time"

// BackupReq asks for a backup of a system.
type BackupReq struct {
	// RepositoryID is the ID or name of the registered repository to write to.
	RepositoryID string `json:"repositoryId" validate:"required,max=255" example:"3f2a9c1e6b7d4a80"`
	// Devices are the partitions to capture. Left out, they are chosen from
	// what the system's agent reported.
	Devices []string `json:"devices,omitempty" validate:"max=128,unique,dive,required,max=4096" example:"/dev/nvme0n1p1"`
	// Compression names the codec, empty for the default.
	Compression string `json:"compression,omitempty" validate:"omitempty,oneof=zstd none" example:"zstd"`
}

// BackupAcceptedResp acknowledges a backup request. The backup itself runs
// later: follow it at the job's address.
type BackupAcceptedResp struct {
	JobID    string `json:"jobId"    example:"5e6f70819a0b1c2d"`
	SystemID string `json:"systemId" example:"1a2b3c4d5e6f7081"`
	Status   string `json:"status"   example:"queued"`
	// Devices are the partitions the job will capture.
	Devices   []string  `json:"devices" example:"/dev/nvme0n1p1"`
	ExpiresAt time.Time `json:"expiresAt"`
	// Warnings are reasons to look twice that did not stop the request, such
	// as an inventory too old to trust.
	Warnings []string `json:"warnings,omitempty"`
}

// BackupInfo is one entry of the backup history: what was asked for and what
// became of it. A backup that failed before producing a set has no BackupSetID.
type BackupInfo struct {
	BackupSetID string `json:"backupSetId,omitempty" example:"9c0d1e2f3a4b5c6d"`
	JobID       string `json:"jobId"                 example:"5e6f70819a0b1c2d"`
	SystemID    string `json:"systemId"              example:"1a2b3c4d5e6f7081"`
	// Status is the job's status.
	Status string `json:"status" example:"completed"`
	// SetStatus is the set's own status as the agent reported it, which can be
	// partial when some of its items failed.
	SetStatus    string   `json:"setStatus,omitempty"    example:"completed"`
	RepositoryID string   `json:"repositoryId,omitempty" example:"3f2a9c1e6b7d4a80"`
	Devices      []string `json:"devices"                example:"/dev/nvme0n1p1"`
	Compression  string   `json:"compression,omitempty"  example:"zstd"`
	// Items is how many backup files the set holds, and Manifest names the
	// file describing them in the repository.
	Items    int    `json:"items"              example:"3"`
	Manifest string `json:"manifest,omitempty" example:"9c0d1e2f3a4b5c6d-manifest.json"`
	// ProcessedBytes is how much was uploaded, compressed.
	ProcessedBytes int64      `json:"processedBytes"       example:"1048576"`
	Message        string     `json:"message,omitempty"    example:"Backup set 9c0d1e2f3a4b5c6d (completed, 3 item(s))"`
	RequestedBy    string     `json:"requestedBy,omitempty" example:"admin"`
	CreatedAt      time.Time  `json:"createdAt"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	FinishedAt     *time.Time `json:"finishedAt,omitempty"`
}

// BackupListResp lists the backup history.
type BackupListResp struct {
	Items []BackupInfo `json:"items"`
}
