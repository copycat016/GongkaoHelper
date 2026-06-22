// Package version exposes build-time metadata.
//
// 默认值在源码中，构建时通过 -ldflags 注入：
//
//	go build -ldflags "-X gkweb/backend/internal/version.Version=v0.1.0-preview \
//	                   -X gkweb/backend/internal/version.Channel=preview \
//	                   -X gkweb/backend/internal/version.Commit=$(git rev-parse --short HEAD)"
package version

var (
	// Version 形如 v0.1.0-preview
	Version = "dev"
	// Channel 形如 stable / preview / dev
	Channel = "dev"
	// Commit 短 commit hash
	Commit = "unknown"
)
