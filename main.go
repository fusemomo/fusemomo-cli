package main

import (
	"github.com/fusemomo/fusemomo-cli/cmd"
)

// Build-time variables injected via ldflags:
//
//	go build -ldflags "-X main.Version=v1.0.0 -X main.Commit=abc1234 -X main.BuiltAt=2026-01-01T00:00:00Z"
var (
	Version = "dev"
	Commit  = "none"
	BuiltAt = "unknown"
)

func main() {
	cmd.Execute(Version, Commit, BuiltAt)
}
