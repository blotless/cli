package cli

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestLineHandler(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(&buf, false)
	log.Info("starting ollama", slog.String("host", "127.0.0.1:54976"))
	log.Info("pulling ollama model", slog.String("model", "qwen2.5-coder:3b"))
	log.Info("walk", slog.String("detail", "12 files"))
	log.Info("stopping ollama", slog.Int("pid", 4242))
	got := buf.String()
	if strings.Contains(got, "level=INFO") || strings.Contains(got, "time=") {
		t.Fatalf("raw slog leaked: %s", got)
	}
	if !strings.Contains(got, "blotless: starting ollama on 127.0.0.1:54976") {
		t.Fatal(got)
	}
	if !strings.Contains(got, "blotless: pulling ollama model qwen2.5-coder:3b") {
		t.Fatal(got)
	}
	if !strings.Contains(got, "blotless: walk  12 files") {
		t.Fatal(got)
	}
	if !strings.Contains(got, "blotless: stopping ollama  pid 4242") {
		t.Fatal(got)
	}
}

func TestLineHandlerEndpoint(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(&buf, false)
	log.Info("rewrite", slog.String("endpoint", "http://127.0.0.1:11434"))
	if !strings.Contains(buf.String(), "at http://127.0.0.1:11434") {
		t.Fatal(buf.String())
	}
}

func TestVerboseLogger(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(&buf, true)
	log.Debug("debug line")
	if !strings.Contains(buf.String(), "debug line") {
		t.Fatal(buf.String())
	}
}

func TestLineHandlerWithAttrsAndGroup(t *testing.T) {
	var buf bytes.Buffer
	h := lineHandler{w: &buf, min: slog.LevelInfo}
	h2 := h.WithAttrs([]slog.Attr{slog.String("k", "v")})
	h3 := h.WithGroup("g")
	if h2.(lineHandler).w != &buf || h3.(lineHandler).w != &buf {
		t.Fatal("handlers should preserve writer")
	}
}
