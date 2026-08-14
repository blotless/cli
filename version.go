package cli

import (
	"encoding/json"
	"fmt"
	goruntime "runtime"
	"runtime/debug"
)

// Build metadata — set via -ldflags at release time.
var (
	// Version is the release tag or "dev".
	Version = "dev"
	// Commit is the short git SHA.
	Commit = "unknown"
	// BuildDate is the UTC build timestamp.
	BuildDate = "unknown"
)

// BuildInfo is the structured version payload for humans and automation.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	CGO       string `json:"cgo"`
}

// CurrentBuildInfo returns runtime + ldflag build metadata.
func CurrentBuildInfo(version string) BuildInfo {
	v := version
	if v == "" {
		v = Version
	}
	commit := Commit
	date := BuildDate
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if commit == "unknown" || commit == "" {
					c := s.Value
					if len(c) > 12 {
						c = c[:12]
					}
					commit = c
				}
			case "vcs.time":
				if date == "unknown" || date == "" {
					date = s.Value
				}
			}
		}
	}
	return BuildInfo{
		Version:   v,
		Commit:    commit,
		BuildDate: date,
		GoVersion: goruntime.Version(),
		OS:        goruntime.GOOS,
		Arch:      goruntime.GOARCH,
		CGO:       "0", // blotless is always built with CGO_ENABLED=0
	}
}

func formatBuildInfo(bi BuildInfo, asJSON bool) string {
	if asJSON {
		b, err := json.MarshalIndent(bi, "", "  ")
		if err != nil {
			return bi.Version + "\n"
		}
		return string(b) + "\n"
	}
	return fmt.Sprintf("blotless %s\n  commit:     %s\n  built:      %s\n  go:         %s\n  platform:   %s/%s\n  cgo:        %s\n",
		bi.Version, bi.Commit, bi.BuildDate, bi.GoVersion, bi.OS, bi.Arch, bi.CGO)
}
