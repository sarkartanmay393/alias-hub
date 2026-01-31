// Package version provides version information for the ah CLI tool.
// The version is set at build time using ldflags.
package version

// Version is the current version of the ah CLI.
// This value should be updated for each release and can be
// overridden at build time with:
//
//	go build -ldflags "-X github.com/sarkartanmay393/ah/pkg/version.Version=x.y.z"
var Version = "1.2.0"
