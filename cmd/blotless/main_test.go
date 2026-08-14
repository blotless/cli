package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/blotless/cli"
)

func TestRunVersion(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"version"}, &out, &errBuf, bytes.NewReader(nil))
	if code != cli.ExitOK {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), cli.Version) {
		t.Fatalf("got %q", out.String())
	}
}

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"--help"}, &out, &bytes.Buffer{}, bytes.NewReader(nil))
	if code != cli.ExitOK {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(out.String(), "inspect") {
		t.Fatalf("%s", out.String())
	}
}
