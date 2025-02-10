package utils

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit_sha": CommitSHA,
		"build_time": BuildTime,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
}

func GetVersionString() string {
	return fmt.Sprintf("%s-%s", Version, CommitSHA)
}
