package model

import "time"

// The types below are the planned contract of the redundancy API (TM-AI-HA-02,
// plan/ha-roadmap.md). Its routes answer 501 until the work behind them exists,
// so these shapes are what a client can plan against, not what it gets today.

// HAResourceSpec is the replicated volume of an HA group.
type HAResourceSpec struct {
	// Name identifies the replication resource. Empty has the framework issue
	// one; set, it adopts a resource that already exists.
	Name string `json:"name,omitempty" validate:"max=63" example:"aiha-r0"`
	// Device is what the filesystem sits on, and BackingDisk the volume holding
	// the data on each node, as a stable by-id path.
	Device      string `json:"device"      validate:"required,max=4096" example:"/dev/drbd0"`
	BackingDisk string `json:"backingDisk" validate:"required,max=4096" example:"/dev/disk/by-id/nvme-Amazon_Elastic_Block_Store_vol0123"`
	MountPoint  string `json:"mountPoint"  validate:"required,max=4096" example:"/data"`
	Port        int    `json:"port"        validate:"required,min=1,max=65535" example:"7789"`
}

// HAServiceSpec is what is health-checked on each node and moved in a switch.
type HAServiceSpec struct {
	Unit       string `json:"unit"                 validate:"required,max=255" example:"ai-service.service"`
	Port       int    `json:"port"                 validate:"required,min=1,max=65535" example:"8000"`
	HealthPath string `json:"healthPath,omitempty" validate:"max=2048" example:"/healthz"`
	// ExpectedGPUs is how many accelerators each node must see. Zero skips
	// the check.
	ExpectedGPUs int `json:"expectedGpus,omitempty" validate:"min=0" example:"1"`
}

// HAGroupReq registers an HA group: an active and a standby system, the volume
// replicated between them and the service that runs on it.
type HAGroupReq struct {
	Name            string         `json:"name"            validate:"required,max=63" example:"inference-ha"`
	PrimarySystemID string         `json:"primarySystemId" validate:"required" example:"1a2b3c4d5e6f7081"`
	StandbySystemID string         `json:"standbySystemId" validate:"required" example:"2b3c4d5e6f708192"`
	Resource        HAResourceSpec `json:"resource"`
	Service         HAServiceSpec  `json:"service"`
	// ServiceAddress is where clients reach the service. It is recorded so
	// that a completed switch can say what to repoint.
	ServiceAddress string `json:"serviceAddress,omitempty" validate:"max=2048" example:"10.0.1.20:8000"`
}

// HAReplicationStatus is the replication state a node reports.
type HAReplicationStatus struct {
	Role           string `json:"role"           example:"Primary"`
	Connection     string `json:"connection"     example:"Connected"`
	Disk           string `json:"disk"           example:"UpToDate"`
	OutOfSyncBytes int64  `json:"outOfSyncBytes" example:"0"`
	Error          string `json:"error,omitempty"`
}

// HANodeStatus is one node's latest health report.
type HANodeStatus struct {
	SystemID string `json:"systemId" example:"1a2b3c4d5e6f7081"`
	// Role is primary or standby.
	Role        string               `json:"role" example:"primary"`
	ReportedAt  *time.Time           `json:"reportedAt,omitempty"`
	ServiceUp   bool                 `json:"serviceUp"   example:"true"`
	PortOpen    bool                 `json:"portOpen"    example:"true"`
	HealthCode  int                  `json:"healthCode,omitempty"      example:"200"`
	LatencyMs   int64                `json:"healthLatencyMs,omitempty" example:"35"`
	GPUsSeen    int                  `json:"gpusSeen,omitempty"        example:"1"`
	Replication *HAReplicationStatus `json:"replication,omitempty"`
}

// HAGroupInfo is an HA group with its state.
type HAGroupInfo struct {
	GroupID string `json:"groupId" example:"6a7b8c9d0e1f2a3b"`
	Name    string `json:"name"    example:"inference-ha"`
	// State is registered, configuring, syncing, protected, degraded,
	// failover_pending, switching or error. It governs what may be done next.
	State string `json:"state" example:"protected"`
	// Verdict is ok, attention or error, derived anew at each evaluation, and
	// Reasons explains it. It is not State: a broken replication link is a
	// verdict of error on a group that is degraded.
	Verdict        string         `json:"verdict"           example:"ok"`
	Reasons        []string       `json:"reasons,omitempty"`
	Primary        HANodeStatus   `json:"primary"`
	Standby        HANodeStatus   `json:"standby"`
	Resource       HAResourceSpec `json:"resource"`
	Service        HAServiceSpec  `json:"service"`
	ServiceAddress string         `json:"serviceAddress,omitempty" example:"10.0.1.20:8000"`
	CreatedAt      time.Time      `json:"createdAt"`
}

// HAGroupListResp lists HA groups.
type HAGroupListResp struct {
	Items []HAGroupInfo `json:"items"`
}

// HAEvent is one change in an HA group's state. Seq only increases, so a
// reader resumes from the last one it saw.
type HAEvent struct {
	Seq     int64     `json:"seq"     example:"42"`
	GroupID string    `json:"groupId" example:"6a7b8c9d0e1f2a3b"`
	Kind    string    `json:"kind"    example:"state_changed"`
	From    string    `json:"from,omitempty"   example:"protected"`
	To      string    `json:"to,omitempty"     example:"degraded"`
	Reason  string    `json:"reason,omitempty" example:"Replication disconnected"`
	At      time.Time `json:"at"`
}

// HAEventListResp lists HA events in sequence order.
type HAEventListResp struct {
	Items []HAEvent `json:"items"`
}

// HAApproval records who authorized an action that cannot be undone. Every
// switch needs one: nothing is promoted on the framework's own judgement.
type HAApproval struct {
	ApprovedBy string `json:"approvedBy" validate:"required,max=255" example:"operator@example.com"`
	Reason     string `json:"reason,omitempty" validate:"max=1024" example:"Planned maintenance"`
	// IncidentID links the action to the incident it answers, if any.
	IncidentID string `json:"incidentId,omitempty" validate:"max=64" example:"7c8d9e0f1a2b3c4d"`
}

// HAFailoverReq approves a failover. ConfirmFenced states the old primary has
// really stopped writing, which only someone outside the framework can know.
// AcceptDataLoss acknowledges a gap in replication, when there is one.
type HAFailoverReq struct {
	HAApproval
	ConfirmFenced  bool `json:"confirmFenced"  validate:"required" example:"true"`
	AcceptDataLoss bool `json:"acceptDataLoss" example:"false"`
}

// HATransitionInfo is one switch or recovery of an HA group.
type HATransitionInfo struct {
	TransitionID string `json:"transitionId" example:"8d9e0f1a2b3c4d5e"`
	GroupID      string `json:"groupId"      example:"6a7b8c9d0e1f2a3b"`
	// Kind is switchover, failover, resume or rejoin.
	Kind string `json:"kind" example:"switchover"`
	// Status is pending, running, completed or failed, and Phase the step it
	// is on.
	Status           string     `json:"status"           example:"running"`
	Phase            string     `json:"phase,omitempty"  example:"demoting"`
	JobIDs           []string   `json:"jobIds,omitempty"`
	RequestedBy      string     `json:"requestedBy,omitempty" example:"admin"`
	ApprovedBy       string     `json:"approvedBy,omitempty"  example:"operator@example.com"`
	IncidentID       string     `json:"incidentId,omitempty"  example:"7c8d9e0f1a2b3c4d"`
	FencingConfirmed bool       `json:"fencingConfirmed" example:"false"`
	DataLossAccepted bool       `json:"dataLossAccepted" example:"false"`
	Message          string     `json:"message,omitempty"`
	StartedAt        time.Time  `json:"startedAt"`
	FinishedAt       *time.Time `json:"finishedAt,omitempty"`
}

// HATransitionListResp lists an HA group's transitions.
type HATransitionListResp struct {
	Items []HATransitionInfo `json:"items"`
}
