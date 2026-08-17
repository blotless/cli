package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := Execute([]string{"--help"}, Options{Stdout: &out, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	got := out.String()
	for _, s := range []string{"inspect", "clean", "rewrite", "rules", "version", "aggressive"} {
		if !strings.Contains(got, s) {
			t.Fatalf("help missing %q:\n%s", s, got)
		}
	}
	if strings.Contains(got, "llm-endpoint") {
		t.Fatalf("LLM flags should be on clean, not root:\n%s", got)
	}
	var inspectOut bytes.Buffer
	code = Execute([]string{"inspect", "--help"}, Options{Stdout: &inspectOut, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("inspect help code=%d err=%s", code, errBuf.String())
	}
	if !strings.Contains(inspectOut.String(), "scan") {
		t.Fatalf("inspect help missing scan alias:\n%s", inspectOut.String())
	}
	var cleanOut bytes.Buffer
	code = Execute([]string{"clean", "--help"}, Options{Stdout: &cleanOut, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("clean help code=%d err=%s", code, errBuf.String())
	}
	got = cleanOut.String()
	for _, s := range []string{"nfkc", "strength", "in-place", "write", "llm-endpoint", "layer-b"} {
		if !strings.Contains(got, s) {
			t.Fatalf("clean help missing %q:\n%s", s, got)
		}
	}
}

func TestVersion(t *testing.T) {
	var out bytes.Buffer
	code := Execute([]string{"version"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "v0.0.0-test"})
	if code != ExitOK {
		t.Fatalf("code=%d", code)
	}
	got := out.String()
	if !strings.Contains(got, "v0.0.0-test") || !strings.Contains(got, "go:") {
		t.Fatalf("got %q", got)
	}
	var jsonOut bytes.Buffer
	code = Execute([]string{"version", "--json"}, Options{Stdout: &jsonOut, Stderr: ioDiscard{}, Version: "v0.0.0-test"})
	if code != ExitOK {
		t.Fatalf("json code=%d", code)
	}
	if !strings.Contains(jsonOut.String(), `"version": "v0.0.0-test"`) {
		t.Fatalf("%s", jsonOut.String())
	}
}

func TestRulesLs(t *testing.T) {
	var out bytes.Buffer
	code := Execute([]string{"rules", "ls"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), "unicode.zwsp") {
		t.Fatalf("missing unicode.zwsp: %s", out.String())
	}
}

func TestRulesExplain(t *testing.T) {
	var out bytes.Buffer
	code := Execute([]string{"rules", "explain", "unicode.zwsp"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), "unicode.zwsp") {
		t.Fatalf("%s", out.String())
	}
	var errBuf bytes.Buffer
	code = Execute([]string{"rules", "explain", "no.such"}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
}

func TestScanFindsZWSP(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("a\u200bb"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"scan", path, "--format", "json", "--fail-on", "certain"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitFound {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), "unicode.zwsp") {
		t.Fatalf("%s", out.String())
	}
	if !strings.Contains(out.String(), `"percent"`) {
		t.Fatalf("missing score: %s", out.String())
	}
}

func TestCleanJSONAfterScore(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("a\u200bb"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"clean", path, "--format", "json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	got := out.String()
	if !strings.Contains(got, `"percent"`) || !strings.Contains(got, `"after"`) {
		t.Fatalf("%s", got)
	}
}

func TestYAMLExcludeSkipsTests(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/a.go", []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testSrc := []byte("package a\n// Co-authored-by: Cursor <cursor@example.com>\n")
	if err := os.WriteFile(dir+"/a_test.go", testSrc, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := []byte("exclude:\n  - '*_test.go'\n")
	if err := os.WriteFile(dir+"/.blotless.yaml", cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"scan", ".", "--format", "json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	got := out.String()
	if strings.Contains(got, "a_test.go") || strings.Contains(got, "stamp.co_authored_by_ai") {
		t.Fatalf("test file not excluded: %s", got)
	}
}

func TestInspectHasNoLLMFlag(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("a\u200bb"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	code := Execute([]string{"inspect", path, "--llm", "ollama"}, Options{Stdout: &out, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s out=%s", code, errBuf.String(), out.String())
	}
	if !strings.Contains(errBuf.String(), "unknown flag") && !strings.Contains(errBuf.String(), "unknown") {
		t.Fatalf("expected unknown flag: %s", errBuf.String())
	}
}

func TestInspectJSONLayers(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("a\u200bb"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"inspect", path, "--json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK && code != ExitFound {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	got := out.String()
	if !strings.Contains(got, `"layers"`) || !strings.Contains(got, `"a"`) {
		t.Fatalf("missing layers: %s", got)
	}
	if !strings.Contains(got, "unicode.zwsp") {
		t.Fatalf("%s", got)
	}
}

func TestLayerBWithoutLLMStillRuns(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/notes.md"
	body := strings.Repeat("This is ordinary prose without Layer A marks. ", 20)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	code := Execute([]string{"clean", path, "--layer-b", "--format", "json"}, Options{Stdout: &out, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s out=%s", code, errBuf.String(), out.String())
	}
	if strings.Contains(errBuf.String(), "starting ollama") {
		t.Fatalf("must not start ollama: %s", errBuf.String())
	}
	if !strings.Contains(out.String(), "statwm.layer_b") && !strings.Contains(out.String(), `"b":`) {
		t.Fatalf("expected Layer B in report: %s", out.String())
	}
}

func TestLayerBASTTransformGo(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/main.go"
	src := "package main\n\nfunc Hello(name string) string {\n\ttmp := name\n\treturn tmp\n}\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"clean", path, "--layer-b", "--format", "json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "statwm.ast_transform") && !strings.Contains(got, "AST transform") {
		t.Fatalf("expected AST transform: %s", got)
	}
}

func TestInspectAlias(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("a\u200bb"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"inspect", path, "--json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK && code != ExitFound {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), "unicode.zwsp") {
		t.Fatalf("%s", out.String())
	}
}

func TestInvalidStrength(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	code := Execute([]string{"clean", path, "--strength", "nope"}, Options{Stdout: &out, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "strength") {
		t.Fatalf("err=%s", errBuf.String())
	}
}

func TestCleanNFKC(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("a\u212ab"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"clean", path, "--nfkc", "--write"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "aKb" {
		t.Fatalf("got %q", got)
	}
}

func TestInspectAggressive(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("sc\u0430n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Execute([]string{"inspect", path, "--aggressive", "--json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK && code != ExitFound {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), "confusable") && !strings.Contains(out.String(), "homoglyph") {
		t.Fatalf("expected aggressive hit: %s", out.String())
	}
}

func TestRewritePrintPrompt(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/n.txt"
	if err := os.WriteFile(path, []byte("Hello world."), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	code := Execute([]string{"rewrite", path, "--backend", "print-prompt"}, Options{Stdout: &out, Stderr: &errBuf, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "token level") || !strings.Contains(out.String(), "Hello world.") {
		t.Fatalf("%s", out.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var errBuf bytes.Buffer
	code := Execute([]string{"nope"}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
