package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type lineHandler struct {
	w   io.Writer
	min slog.Level
}

func (h lineHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.min
}

func (h lineHandler) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message
	var host, model, endpoint, detail, pid string
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "host":
			host = a.Value.String()
		case "model":
			model = a.Value.String()
		case "endpoint":
			endpoint = a.Value.String()
		case "detail":
			detail = a.Value.String()
		case "pid":
			pid = a.Value.String()
		}
		return true
	})
	switch {
	case host != "":
		msg = msg + " on " + host
	case model != "":
		msg = msg + " " + model
	case endpoint != "":
		msg = msg + " at " + endpoint
	case detail != "":
		msg = msg + "  " + detail
	case pid != "":
		msg = msg + "  pid " + pid
	}
	_, err := fmt.Fprintf(h.w, "blotless: %s\n", msg)
	return err
}

func (h lineHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h lineHandler) WithGroup(string) slog.Handler { return h }

func newLogger(w io.Writer, verbose bool) *slog.Logger {
	if verbose {
		return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(lineHandler{w: w, min: slog.LevelInfo})
}
