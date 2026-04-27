// Package version is set at link time via -ldflags, with runtime fallback from build metadata.
package version

import (
	"regexp"
	"runtime/debug"
	"strings"
)

// tagSemver matches a release line from GoReleaser / make VERSION=v1.2.3 (no pre-release suffix).
var tagSemver = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

var (
	// Name is the full product name (readme §3).
	Name = "Docker Quick-ops"
	// Version is the release tag, module version from go install, or "dev".
	Version = "dev"
	// Commit is a short VCS revision, or "local" when unavailable.
	Commit = ""
)

func init() {
	enrichFromBuildInfo()
}

func enrichFromBuildInfo() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		normalizeCommit()
		normalizeVersion()
		return
	}

	needCommit := Commit == "" || Commit == "unknown"
	if needCommit {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				rev := s.Value
				if len(rev) > 7 {
					rev = rev[:7]
				}
				Commit = rev
				break
			}
		}
	}
	normalizeCommit()

	needVersion := Version == "" || Version == "unknown"
	if needVersion {
		Version = "dev"
	}
	if Version == "dev" {
		mv := strings.TrimSpace(info.Main.Version)
		if mv != "" && mv != "(devel)" {
			Version = mv
		}
	}
	normalizeVersion()
}

func normalizeCommit() {
	if Commit == "" || Commit == "unknown" {
		Commit = "local"
	}
}

func normalizeVersion() {
	if Version == "" || Version == "unknown" {
		Version = "dev"
	}
}

// ShowModuleVersionNote is true when the reported version is not a plain v1.2.3 from release
// ldflags — e.g. go install reported pseudo-version. Then dq version prints a short extra line.
func ShowModuleVersionNote() bool {
	v := strings.TrimSpace(Version)
	if v == "" || v == "dev" {
		return false
	}
	return !tagSemver.MatchString(v)
}
