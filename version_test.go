package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCurrentBuildInfo(t *testing.T) {
	bi := CurrentBuildInfo("")
	if bi.Version != Version {
		t.Fatalf("version=%q", bi.Version)
	}
	if bi.GoVersion == "" || bi.OS == "" || bi.Arch == "" {
		t.Fatalf("%+v", bi)
	}
	if bi.CGO != "0" {
		t.Fatalf("cgo=%q", bi.CGO)
	}
	bi2 := CurrentBuildInfo("custom")
	if bi2.Version != "custom" {
		t.Fatalf("custom version=%q", bi2.Version)
	}
}

func TestFormatBuildInfoJSON(t *testing.T) {
	bi := BuildInfo{Version: "v1", Commit: "abc", BuildDate: "now", GoVersion: "go1.26", OS: "darwin", Arch: "arm64", CGO: "0"}
	got := formatBuildInfo(bi, true)
	if !json.Valid([]byte(strings.TrimSpace(got))) {
		t.Fatalf("%q", got)
	}
	if !strings.Contains(got, `"version": "v1"`) {
		t.Fatalf("%q", got)
	}
}

func TestFormatBuildInfoText(t *testing.T) {
	bi := CurrentBuildInfo("v-test")
	got := formatBuildInfo(bi, false)
	if !strings.Contains(got, "blotless v-test") || !strings.Contains(got, "platform:") {
		t.Fatalf("%q", got)
	}
}
