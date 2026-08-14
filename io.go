package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MaxInputBytes refuses oversized rewrite/single-file inputs (WR parity).
const MaxInputBytes = 8 << 20

// ReadTextInput reads a file path or "-" for stdin.
func ReadTextInput(path string, stdin io.Reader, forceText bool) ([]byte, error) {
	var raw []byte
	var err error
	if path == "" || path == "-" {
		raw, err = io.ReadAll(io.LimitReader(stdin, MaxInputBytes+1))
		if err != nil {
			return nil, err
		}
	} else {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return nil, fmt.Errorf("expected a file, got directory: %s", path)
		}
		if info.Size() > MaxInputBytes {
			return nil, fmt.Errorf("refusing input larger than %d bytes: %s", MaxInputBytes, path)
		}
		raw, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	if int64(len(raw)) > MaxInputBytes {
		return nil, fmt.Errorf("refusing input larger than %d bytes", MaxInputBytes)
	}
	if !forceText && looksBinary(path, raw) {
		return nil, fmt.Errorf("input looks like a binary container; pass --force-text to override (will corrupt binary formats)")
	}
	return raw, nil
}

func looksBinary(path string, raw []byte) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp", ".pdf", ".docx", ".odt":
		return true
	}
	if bytes.HasPrefix(raw, []byte{0x89, 'P', 'N', 'G'}) || bytes.HasPrefix(raw, []byte{0xff, 0xd8}) || bytes.HasPrefix(raw, []byte("%PDF")) {
		return true
	}
	if len(raw) >= 12 && bytes.Equal(raw[:4], []byte("RIFF")) && bytes.Equal(raw[8:12], []byte("WEBP")) {
		return true
	}
	if bytes.HasPrefix(raw, []byte("PK")) && (bytes.Contains(raw[:min(len(raw), 64<<10)], []byte("word/document.xml")) ||
		bytes.Contains(raw[:min(len(raw), 64<<10)], []byte("content.xml"))) {
		return true
	}
	n := len(raw)
	if n > 8192 {
		n = 8192
	}
	return bytes.IndexByte(raw[:n], 0) >= 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DerivedPath returns path with an inserted suffix before the extension (WR cleaned_path).
func DerivedPath(path, suffix string) string {
	if path == "" || path == "-" {
		return ""
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	if suffix == "" {
		suffix = ".cleaned"
	}
	return base + suffix + ext
}

// BackupPath copies src to src+suffix (default .bak) and returns the backup path.
func BackupPath(src, suffix string) (string, error) {
	if suffix == "" {
		suffix = ".bak"
	}
	dst := src + suffix
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return "", err
	}
	return dst, nil
}

// WriteTextOutput writes to path, or stdout when path is empty/"-".
func WriteTextOutput(w io.Writer, path string, data []byte) error {
	if path == "" || path == "-" {
		if _, err := w.Write(data); err != nil {
			return err
		}
		if len(data) == 0 || data[len(data)-1] != '\n' {
			_, err := io.WriteString(w, "\n")
			return err
		}
		return nil
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

// ForceKind is --as text|image|container|auto.
func ForceKind(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return "auto", nil
	case "text", "image", "container":
		return strings.ToLower(strings.TrimSpace(s)), nil
	default:
		return "", fmt.Errorf("--as must be auto|text|image|container")
	}
}
