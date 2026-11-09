// Package model defines the request and response types exchanged by the AI-HA
// REST API.
package model

// SimpleMsg is the generic response body used for status and error replies.
type SimpleMsg struct {
	Message string `json:"message" example:"Service ready"`
	// Code identifies a failure for programs to branch on; Message is for
	// people and may be reworded. The management API sets it on the failures
	// it answers itself; one turned away earlier, such as by the rate limiter,
	// has only a message and its HTTP status.
	Code string `json:"code,omitempty" example:"NOT_FOUND"`
}

// Failure codes of the management API. They are part of the contract: unlike
// a message, a code keeps its meaning across releases.
const (
	CodeInvalidRequest      = "INVALID_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeInternal            = "INTERNAL_ERROR"
	CodeUnavailable         = "SERVICE_UNAVAILABLE"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
	CodeRepositoryNotFound  = "REPOSITORY_NOT_FOUND"
	CodeRepositoryExists    = "REPOSITORY_EXISTS"
	CodeBackupSetNotFound   = "BACKUP_SET_NOT_FOUND"
	CodeBackupSetAmbiguous  = "BACKUP_SET_AMBIGUOUS"
	CodeBackupSetUnusable   = "BACKUP_SET_UNUSABLE"
	CodeNoInventory         = "NO_INVENTORY"
	CodeInventoryIncomplete = "INVENTORY_INCOMPLETE"
	CodeDeviceNotBackupable = "DEVICE_NOT_BACKUPABLE"
	CodeJobFinished         = "JOB_FINISHED"
)

// FrameworkAccess is how an agent reaches and verifies this framework. It goes
// out with a registration token or a restore job, whose receiver has to
// configure an agent with it.
type FrameworkAccess struct {
	URL string `json:"url" example:"https://framework.example.com:8443"`
	// CertPin says the agent pins the certificate named by CertSHA256, for a
	// self-signed certificate. False means it verifies the certificate against
	// its CA bundle, and CertSHA256 is empty.
	CertPin    bool   `json:"certPin"              example:"true"`
	CertSHA256 string `json:"certSha256,omitempty" example:"3911ba8d8fe5db35ae90469636dcd464c8b7aee357dc1bf9f9e6d8db3b77e81e"`
}
