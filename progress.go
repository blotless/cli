package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/blotless/engine"
	"github.com/blotless/engine/domain"
	"github.com/blotless/engine/ports"
)

const barWidth = 24

type liveUI struct {
	out, err io.Writer
	tty      bool
	mu       sync.Mutex
	lastFile string
	lastBar  int
	shown    int
	done     int
	total    int
	path     string
	detector string
}

func newLiveUI(out, err io.Writer) *liveUI {
	return &liveUI{
		out: out,
		err: err,
		tty: isTTY(err),
	}
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func (l *liveUI) File(done, total int, path, detector string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.done, l.total, l.path, l.detector = done, total, path, detector
	if l.tty {
		l.draw()
		return
	}
	if done == l.shown && !strings.Contains(detector, "heuristic") {
		return
	}
	l.shown = done
	fmt.Fprintf(l.err, "blotless: %d/%d  %s  %s\n", done, total, detector, shortPath(path, 40))
}

func (l *liveUI) Finding(f domain.Finding) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clear()
	l.lastFile = engine.WriteFinding(l.out, f, l.lastFile)
	if l.tty {
		l.draw()
	}
}

func (l *liveUI) Finish() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clear()
	l.lastBar = 0
}

func (l *liveUI) draw() {
	if l.total < 1 {
		l.total = 1
	}
	pct := l.done * 100 / l.total
	filled := l.done * barWidth / l.total
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	line := fmt.Sprintf("\r%s [%s] %d/%d  %d%%  %s  %s",
		dim("blotless"), bar, l.done, l.total, pct, l.detector, shortPath(l.path, 32))
	if n := utf8.RuneCountInString(line); n > l.lastBar {
		l.lastBar = n
	}
	fmt.Fprint(l.err, pad(line, l.lastBar))
	l.shown = l.done
}

func (l *liveUI) clear() {
	if !l.tty || l.lastBar == 0 {
		return
	}
	fmt.Fprint(l.err, "\r"+strings.Repeat(" ", l.lastBar)+"\r")
}

func shortPath(p string, max int) string {
	p = filepath.ToSlash(p)
	if utf8.RuneCountInString(p) <= max {
		return p
	}
	runes := []rune(p)
	return "\u2026" + string(runes[len(runes)-(max-1):])
}

func pad(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func dim(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

var _ ports.Progress = (*liveUI)(nil)
