package model

import "time"

// RepositoryInfo describes a registered backup repository without its
// credentials. It is what listing and inspecting a repository return;
// RepositoryAccess, which carries the credentials, goes only to an agent
// carrying out a job against the repository.
type RepositoryInfo struct {
	ID       string `json:"id"                 example:"3f2a9c1e6b7d4a80"`
	Name     string `json:"name"               example:"minio-main"`
	Type     string `json:"type"               example:"s3"`
	Provider string `json:"provider"           example:"Minio"`
	Endpoint string `json:"endpoint,omitempty" example:"https://storage.example.com:9000"`
	Region   string `json:"region,omitempty"   example:"us-east-1"`
	Bucket   string `json:"bucket"             example:"aiha-backup"`
	Prefix   string `json:"prefix,omitempty"   example:"prod"`
	// HasSessionToken says whether the credentials include a session token,
	// which expires and is not renewed by the framework.
	HasSessionToken      bool      `json:"hasSessionToken"`
	CreatedAt            time.Time `json:"createdAt"`
	CredentialsUpdatedAt time.Time `json:"credentialsUpdatedAt"`
}

// RepositoryListResp lists registered repositories.
type RepositoryListResp struct {
	Items []RepositoryInfo `json:"items"`
}

// RepositoryReq registers a backup repository. The location cannot change once
// registered: backup sets already written there would be looked for somewhere
// else.
type RepositoryReq struct {
	Name string `json:"name" validate:"required,max=63" example:"minio-main"`
	// Provider is AWS or Minio.
	Provider string `json:"provider" validate:"required,oneof=AWS Minio" example:"Minio"`
	// Endpoint is the URL of an S3-compatible service, empty for AWS.
	Endpoint string `json:"endpoint,omitempty" validate:"max=2048" example:"https://storage.example.com:9000"`
	// Region is required for AWS.
	Region string `json:"region,omitempty" validate:"max=64"   example:"us-east-1"`
	Bucket string `json:"bucket"           validate:"required,max=63" example:"aiha-backup"`
	Prefix string `json:"prefix,omitempty" validate:"max=1024" example:"prod"`
	RepositoryCredentialsReq
}

// RepositoryCredentialsReq carries repository credentials. They are stored
// encrypted and returned by no route.
type RepositoryCredentialsReq struct {
	AccessKey string `json:"accessKey" validate:"required,max=1024" example:"<access-key>"`
	SecretKey string `json:"secretKey" validate:"required,max=1024" example:"<secret-key>"`
	// SessionToken makes the credentials temporary. It is not renewed by the
	// framework: replace the credentials when it expires.
	SessionToken string `json:"sessionToken,omitempty" validate:"max=4096" example:"<session-token>"`
}

// RepositoryVerifyReq asks for a repository's reachability to be checked.
type RepositoryVerifyReq struct {
	// Verifier is who checks: "framework", or "agent:<systemId>" to learn
	// whether that agent can really write there. Only the second predicts
	// whether a backup will succeed.
	Verifier string `json:"verifier" validate:"required" example:"agent:1a2b3c4d5e6f7081"`
}
