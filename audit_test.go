package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blotless/engine"
	"github.com/blotless/engine/domain"
	"github.com/blotless/engine/webaudit"
)

func TestReportHasActionable(t *testing.T) {
	if reportHasActionable(&engine.Result{Findings: []domain.Finding{{Confidence: domain.ConfidenceHeuristic}}}) {
		t.Fatal("possible should not be actionable")
	}
	if !reportHasActionable(&engine.Result{Findings: []domain.Finding{{Confidence: domain.ConfidenceLikely}}}) {
		t.Fatal("likely should be actionable")
	}
	if !reportHasActionable(&engine.Result{Findings: []domain.Finding{{Confidence: domain.ConfidenceCertain}}}) {
		t.Fatal("certain should be actionable")
	}
}

func TestAuditWebMissingTarget(t *testing.T) {
	var errBuf bytes.Buffer
	code := Execute([]string{"audit", "web"}, Options{Stdout: ioDiscard{}, Stderr: &errBuf, Version: "test"})
	if code != ExitErr {
		t.Fatalf("code=%d err=%s", code, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "sitemap") && !strings.Contains(errBuf.String(), "base") {
		t.Fatalf("err=%s", errBuf.String())
	}
}

func TestAuditWebStubbedTable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("a\u200bb"))
	}))
	t.Cleanup(srv.Close)

	oldCollect := testAuditCollect
	oldFetch := testAuditFetch
	t.Cleanup(func() {
		testAuditCollect = oldCollect
		testAuditFetch = oldFetch
	})
	page := srv.URL + "/page.txt"
	testAuditCollect = func(_ context.Context, _ webaudit.Options) (string, []string, error) {
		return srv.URL + "/sitemap.xml", []string{page}, nil
	}
	testAuditFetch = func(rawURL string) ([]byte, string, error) {
		if rawURL != page {
			t.Fatalf("unexpected url %q", rawURL)
		}
		return []byte("a\u200bb"), "text/plain", nil
	}

	var out bytes.Buffer
	code := Execute([]string{"audit", "web", "--sitemap", srv.URL + "/sitemap.xml"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitFound {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "Sitemap:") || !strings.Contains(got, "Findings:") {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(got, "unicode.zwsp") {
		t.Fatalf("missing finding: %s", got)
	}
}

func TestAuditWebStubbedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("clean text"))
	}))
	t.Cleanup(srv.Close)

	oldCollect := testAuditCollect
	oldFetch := testAuditFetch
	t.Cleanup(func() {
		testAuditCollect = oldCollect
		testAuditFetch = oldFetch
	})
	page := srv.URL + "/ok.txt"
	testAuditCollect = func(_ context.Context, _ webaudit.Options) (string, []string, error) {
		return srv.URL + "/sitemap.xml", []string{page}, nil
	}
	testAuditFetch = func(string) ([]byte, string, error) {
		return []byte("clean text"), "text/plain", nil
	}

	var out bytes.Buffer
	code := Execute([]string{"audit", "web", "--sitemap", srv.URL + "/sitemap.xml", "--json"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), `"urls_scanned"`) || !strings.Contains(out.String(), `"findings"`) {
		t.Fatalf("%s", out.String())
	}
}

func TestAuditWebFetchFailure(t *testing.T) {
	oldCollect := testAuditCollect
	oldFetch := testAuditFetch
	t.Cleanup(func() {
		testAuditCollect = oldCollect
		testAuditFetch = oldFetch
	})
	testAuditCollect = func(_ context.Context, _ webaudit.Options) (string, []string, error) {
		return "https://example.com/sitemap.xml", []string{"https://example.com/missing.txt"}, nil
	}
	testAuditFetch = func(string) ([]byte, string, error) {
		return nil, "", context.DeadlineExceeded
	}

	var out bytes.Buffer
	code := Execute([]string{"audit", "web", "--sitemap", "https://example.com/sitemap.xml"}, Options{Stdout: &out, Stderr: ioDiscard{}, Version: "test"})
	if code != ExitOK {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if !strings.Contains(out.String(), "[error]") {
		t.Fatalf("%s", out.String())
	}
}
