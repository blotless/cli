package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/blotless/engine/domain"
	"github.com/spf13/viper"
)

func TestSplitCSVAndCleanStrings(t *testing.T) {
	if splitCSV("") != nil {
		t.Fatal("empty")
	}
	if got := splitCSV(" a, ,b "); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("%v", got)
	}
	if got := cleanStrings([]string{" x ", "", "y"}); len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("%v", got)
	}
}

func TestParseViperStrings(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("include", []string{"a.go", " b.go "})
	if got := parseViperStrings("include"); len(got) != 2 || got[1] != "b.go" {
		t.Fatalf("%v", got)
	}
	viper.Set("exclude", "one,two")
	if got := parseViperStrings("exclude"); len(got) != 2 {
		t.Fatalf("%v", got)
	}
	viper.Set("disable-rule", []any{"r1", " r2 "})
	if got := parseViperStrings("disable-rule"); len(got) != 2 || got[1] != "r2" {
		t.Fatalf("%v", got)
	}
}

func TestExitErrorString(t *testing.T) {
	if got := (exitError{code: ExitFound, msg: "found issues"}).Error(); got != "found issues" {
		t.Fatalf("%q", got)
	}
}

func TestExecuteExitErrorWithMessage(t *testing.T) {
	var errBuf bytes.Buffer
	code := Execute([]string{"version"}, Options{
		Stdout:  ioDiscard{},
		Stderr:  &errBuf,
		Version: "test",
	})
	if code != ExitOK {
		t.Fatalf("code=%d", code)
	}
	// Force a path that prints exitError.msg via cobra — use invalid config file.
	code = Execute([]string{"inspect", t.TempDir(), "--config", t.TempDir() + "/missing.yaml"}, Options{
		Stdout:  ioDiscard{},
		Stderr:  &errBuf,
		Version: "test",
	})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
}

func TestBarOnlyFinding(t *testing.T) {
	var out bytes.Buffer
	b := barOnly{newLiveUI(&out, ioDiscard{})}
	b.Finding(domain.Finding{
		RuleID:     "unicode.zwsp",
		Confidence: domain.ConfidenceCertain,
		Snippet:    "a<U+200B>b",
		Span:       domain.Span{File: "x.txt", Line: 1, Col: 1, Start: 0, End: 1},
	})
}

func TestRulesFormats(t *testing.T) {
	var out bytes.Buffer
	code := Execute([]string{"rules", "ls", "--format", "json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d", code)
	}
	if !json.Valid(out.Bytes()) {
		t.Fatalf("%s", out.String())
	}
	out.Reset()
	code = Execute([]string{"rules", "ls", "--format", "yaml"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(out.String(), "unicode.zwsp") {
		t.Fatalf("%s", out.String())
	}
	out.Reset()
	code = Execute([]string{"rules", "explain", "unicode.zwsp", "--format", "yaml"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(out.String(), "unicode.zwsp:") && !strings.Contains(out.String(), "id:") {
		t.Fatalf("%s", out.String())
	}
}

func TestIncludeExcludeFlags(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/keep.go", []byte("package keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/skip.go", []byte("package skip\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"inspect", dir, "--include", "*.go", "--exclude", "skip.go", "--format", "json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if strings.Contains(out.String(), "skip.go") {
		t.Fatalf("exclude failed: %s", out.String())
	}
}

func TestLoadConfigExplicitMissing(t *testing.T) {
	dir := t.TempDir()
	cfg := dir + "/nope.yaml"
	var errBuf bytes.Buffer
	code := Execute([]string{"inspect", dir, "--config", cfg}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
}

func TestInspectInvalidAsKind(t *testing.T) {
	var errBuf bytes.Buffer
	code := Execute([]string{"inspect", ".", "--as", "nope"}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
}

func TestVersionPersistentJSON(t *testing.T) {
	var out bytes.Buffer
	code := Execute([]string{"--json", "version"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "v-json"})
	if code != ExitOK {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(out.String(), `"version": "v-json"`) {
		t.Fatalf("%s", out.String())
	}
}

func TestCleanWriteInPlaceAlias(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("plain"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"clean", path, "--in-place", "--write"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
}

func TestRewriteStdinAndStats(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := Execute([]string{"rewrite", "-", "--backend", "print-prompt", "--stats"}, Options{
		Stdout:  &out,
		Stderr:  &errBuf,
		Stdin:   strings.NewReader("Hello world."),
		Version: "test",
	})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "Hello world.") {
		t.Fatalf("%s", out.String())
	}
	var stats map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(errBuf.String())), &stats); err != nil {
		t.Fatalf("stats: %v body=%s", err, errBuf.String())
	}
}

func TestRewriteInPlace(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/draft.txt"
	if err := os.WriteFile(path, []byte("Hello world."), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	code := Execute([]string{"rewrite", path, "--backend", "print-prompt", "--in-place", "--backup", ".orig"}, Options{
		Stdout:  &out,
		Stderr:  &errBuf,
		Version: "test",
	})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if _, err := os.Stat(path + ".orig"); err != nil {
		t.Fatalf("backup: %v", err)
	}
}

func TestRewriteInPlaceRequiresPath(t *testing.T) {
	var errBuf bytes.Buffer
	code := Execute([]string{"rewrite", "-", "--backend", "print-prompt", "--in-place"}, Options{
		Stdout:  ioDiscard{},
		Stderr:  &errBuf,
		Stdin:   strings.NewReader("x"),
		Version: "test",
	})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
}

func TestRewriteInvalidBackend(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	var errBuf bytes.Buffer
	code := Execute([]string{"rewrite", path, "--backend", "nope"}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
}

func TestRewriteOllamaHTTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"message":{"content":"rewritten text"}}`)
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("Hello world."), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	code := Execute([]string{
		"rewrite", path, "--backend", "ollama", "--model", "m", "--endpoint", srv.URL, "-o", dir + "/out.txt",
	}, Options{Stdout: &out, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s out=%s", code, errBuf.String(), out.String())
	}
	got, err := os.ReadFile(dir + "/out.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "rewritten") {
		t.Fatalf("%q", got)
	}
}

func TestRewriteDerivedOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"message":{"content":"done"}}`)
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	path := dir + "/draft.md"
	if err := os.WriteFile(path, []byte("Hello world."), 0o644); err != nil {
		t.Fatal(err)
	}
	var errBuf bytes.Buffer
	code := Execute([]string{
		"rewrite", path, "--backend", "ollama", "--model", "m", "--endpoint", srv.URL,
	}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if _, err := os.Stat(dir + "/draft.rewritten.md"); err != nil {
		t.Fatalf("derived output: %v", err)
	}
}

func TestVerboseLogging(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	var errBuf bytes.Buffer
	code := Execute([]string{"inspect", path, "-v"}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "level=DEBUG") && !strings.Contains(errBuf.String(), "time=") {
		// verbose uses text slog handler
		if !strings.Contains(errBuf.String(), "level=") {
			t.Fatalf("expected verbose logs: %s", errBuf.String())
		}
	}
}
