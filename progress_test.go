package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/blotless/engine/domain"
)

func TestLiveFindingStreams(t *testing.T) {
	var out, errBuf bytes.Buffer
	ui := newLiveUI(&out, &errBuf)
	ui.File(1, 2, "a.go", "unicode")
	ui.Finding(domain.Finding{
		RuleID:     "unicode.zwsp",
		Confidence: domain.ConfidenceCertain,
		Snippet:    "hello<U+200B>",
		Span:       domain.Span{File: "a.go", Line: 3, Col: 1, Start: 0, End: 3},
	})
	ui.Finish()
	got := out.String()
	if !strings.Contains(got, "a.go") || !strings.Contains(got, "unicode.zwsp") {
		t.Fatal(got)
	}
	if !strings.Contains(errBuf.String(), "1/2") {
		t.Fatal(errBuf.String())
	}
}

func TestShortPath(t *testing.T) {
	if got := shortPath("engine/engine.go", 40); got != "engine/engine.go" {
		t.Fatal(got)
	}
	got := shortPath(strings.Repeat("a", 50), 10)
	if !strings.HasPrefix(got, "\u2026") || len([]rune(got)) != 10 {
		t.Fatalf("%q", got)
	}
}

func TestPadAndDim(t *testing.T) {
	if pad("abc", 5) != "abc  " {
		t.Fatalf("%q", pad("abc", 5))
	}
	if pad("abcdef", 3) != "abcdef" {
		t.Fatalf("%q", pad("abcdef", 3))
	}
	t.Setenv("NO_COLOR", "1")
	if dim("x") != "x" {
		t.Fatal("NO_COLOR")
	}
	t.Setenv("NO_COLOR", "")
	if dim("x") == "x" {
		t.Fatal("expected ansi dim")
	}
}

func TestLiveUIDrawTTY(t *testing.T) {
	f, err := os.Open("/dev/tty")
	if err != nil {
		t.Skip("no /dev/tty")
	}
	defer f.Close()
	var out bytes.Buffer
	ui := newLiveUI(&out, f)
	ui.File(2, 5, strings.Repeat("p", 60), "unicode")
	ui.Finding(domain.Finding{
		RuleID:     "unicode.zwsp",
		Confidence: domain.ConfidenceCertain,
		Snippet:    "x",
		Span:       domain.Span{File: "a.go", Line: 1, Col: 1, Start: 0, End: 1},
	})
	ui.Finish()
	if out.String() == "" {
		t.Fatal("expected finding output")
	}
}
