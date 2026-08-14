package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDerivedPath(t *testing.T) {
	got := DerivedPath("draft.md", ".rewritten")
	if got != "draft.rewritten.md" {
		t.Fatalf("%q", got)
	}
	if DerivedPath("-", ".x") != "" {
		t.Fatal("dash path")
	}
	if DerivedPath("file", "") != "file.cleaned" {
		t.Fatalf("%q", DerivedPath("file", ""))
	}
}

func TestBackupAndWrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(src, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	bak, err := BackupPath(src, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(bak, ".bak") {
		t.Fatal(bak)
	}
	out := DerivedPath(src, ".cleaned")
	if err := WriteTextOutput(nil, out, []byte("bye")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "bye" {
		t.Fatalf("%q", got)
	}
}

func TestWriteTextOutputStdout(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTextOutput(&buf, "-", []byte("line")); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "line\n" {
		t.Fatalf("%q", buf.String())
	}
	buf.Reset()
	if err := WriteTextOutput(&buf, "-", []byte("already\n")); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "already\n" {
		t.Fatalf("%q", buf.String())
	}
}

func TestReadTextInputFileErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := ReadTextInput(dir, nil, false); err == nil {
		t.Fatal("expected directory error")
	}
	png := filepath.Join(dir, "x.png")
	if err := os.WriteFile(png, []byte{0x89, 'P', 'N', 'G'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTextInput(png, nil, false); err == nil {
		t.Fatal("expected binary error")
	}
	if _, err := ReadTextInput(png, nil, true); err != nil {
		t.Fatal(err)
	}
}

func TestLooksBinary(t *testing.T) {
	if !looksBinary("x.png", []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatal("png magic")
	}
	if !looksBinary("x.bin", []byte("hello\x00world")) {
		t.Fatal("nul byte")
	}
	if looksBinary("x.txt", []byte("plain text")) {
		t.Fatal("plain text")
	}
	if !looksBinary("x.pdf", []byte("%PDF-1.4")) {
		t.Fatal("pdf magic")
	}
}

func TestMin(t *testing.T) {
	if min(1, 2) != 1 || min(3, 2) != 2 {
		t.Fatal("min")
	}
}

func TestForceKind(t *testing.T) {
	k, err := ForceKind("TEXT")
	if err != nil || k != "text" {
		t.Fatalf("%q %v", k, err)
	}
	if _, err := ForceKind("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestReadTextInputStdin(t *testing.T) {
	raw, err := ReadTextInput("-", strings.NewReader("hello"), false)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "hello" {
		t.Fatalf("%q", raw)
	}
}
