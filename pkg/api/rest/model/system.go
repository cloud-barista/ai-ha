package model

import "time"

// ProtectedSystem is a machine under protection. It is known through the agent
// enrolled for it: the framework never reaches into a system itself.
type ProtectedSystem struct {
	SystemID     string `json:"systemId"               example:"1a2b3c4d5e6f7081"`
	Hostname     string `json:"hostname,omitempty"     example:"web01"`
	AgentID      string `json:"agentId"                example:"9b1c2d3e4f5a6b7c"`
	AgentVersion string `json:"agentVersion,omitempty" example:"aiha-agent 0.1.0"`
	// AgentStatus is online, offline or revoked. Online says the agent is
	// reachable, not that the system's work is healthy: an agent can be alive
	// while its job is stuck.
	AgentStatus  string     `json:"agentStatus"          example:"online"`
	RegisteredAt time.Time  `json:"registeredAt"`
	LastSeenAt   *time.Time `json:"lastSeenAt,omitempty"`
}

// ProtectedSystemListResp lists protected systems.
type ProtectedSystemListResp struct {
	Items []ProtectedSystem `json:"items"`
}

// InventorySummary says how current a system's inventory is. CollectedAt is
// when the agent read the system and ReportedAt when the framework heard it: an
// agent reporting punctually from a stale reading is a different fault from one
// that has gone quiet.
type InventorySummary struct {
	CollectedAt time.Time `json:"collectedAt"`
	ReportedAt  time.Time `json:"reportedAt"`
	// Complete is false when the agent could not account for every filesystem
	// the system needs. Backup targets are then not chosen automatically.
	Complete bool `json:"complete" example:"true"`
	// Stale says the reading is old enough that the system may have changed.
	Stale bool `json:"stale" example:"false"`
}

// ProtectedSystemDetail is one protected system. Inventory is absent until the
// agent first reports.
type ProtectedSystemDetail struct {
	ProtectedSystem
	Inventory *InventorySummary `json:"inventory,omitempty"`
}

// SystemInventoryInfo is what a system's agent last reported it is made of.
type SystemInventoryInfo struct {
	SystemID string `json:"systemId" example:"1a2b3c4d5e6f7081"`
	AgentID  string `json:"agentId"  example:"9b1c2d3e4f5a6b7c"`
	InventorySummary
	// AgeSeconds is how long ago the agent read the system, by the framework's
	// clock.
	AgeSeconds int64 `json:"ageSeconds" example:"1200"`
	// Report is the inventory as the agent sent it.
	Report AgentInventoryReq `json:"report"`
}

// BackupTargetsReq names what a system's backups capture.
type BackupTargetsReq struct {
	Devices []string `json:"devices" validate:"required,min=1,max=128,unique,dive,required,max=4096" example:"/dev/nvme0n1p1"`
}
