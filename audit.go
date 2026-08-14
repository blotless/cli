package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/blotless/engine"
	"github.com/blotless/engine/domain"
	"github.com/blotless/engine/webaudit"
	"github.com/spf13/cobra"
)

var (
	auditCollect    = webaudit.Collect
	auditFetchURL   = webaudit.FetchURL
	testAuditCollect func(context.Context, webaudit.Options) (string, []string, error)
	testAuditFetch   func(string) ([]byte, string, error)
)

func (rt *runtime) auditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Extra audits (website sitemap)",
	}
	cmd.AddCommand(rt.auditWebCmd())
	return cmd
}

func (rt *runtime) auditWebCmd() *cobra.Command {
	var sitemap, base string
	var maxPages, maxBytes int
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Audit URLs listed in a sitemap (SSRF-safe)",
		Long: `Download sitemap URLs and run the same inspect pipeline as local files.

Only public HTTP(S) hosts are allowed. Cross-origin redirects are refused.
Optional tools (c2patool/exiftool) are not invoked for remote URLs.`,
		Example: `  blotless audit web --sitemap https://example.com/sitemap.xml --json
  blotless audit web --base https://example.com --max-pages 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			var sm string
			var urls []string
			var err error
			opt := webaudit.Options{
				Sitemap:  sitemap,
				Base:     base,
				MaxPages: maxPages,
				Timeout:  timeout,
				MaxBytes: maxBytes,
			}
			if testAuditCollect != nil {
				sm, urls, err = testAuditCollect(ctx, opt)
			} else {
				sm, urls, err = auditCollect(ctx, opt)
			}
			if err != nil {
				return err
			}
			dir, err := os.MkdirTemp("", "blotless-web-*")
			if err != nil {
				return err
			}
			defer func() { _ = os.RemoveAll(dir) }()

			report := webaudit.Report{Sitemap: sm, Base: base, URLsCollected: len(urls)}
			var paths []string
			urlByPath := map[string]string{}
			for _, u := range urls {
				var data []byte
				var ct string
				if testAuditFetch != nil {
					data, ct, err = testAuditFetch(u)
				} else {
					data, ct, err = auditFetchURL(ctx, nil, u, timeout, maxBytes, nil)
				}
				if err != nil {
					report.Failures = append(report.Failures, webaudit.Failure{URL: u, Error: err.Error()})
					continue
				}
				kind := webaudit.GuessKind(u, data, ct)
				name := filepath.Join(dir, fmt.Sprintf("%d%s", len(paths), webaudit.ExtForKind(kind)))
				if err := os.WriteFile(name, data, 0o600); err != nil {
					report.Failures = append(report.Failures, webaudit.Failure{URL: u, Error: err.Error()})
					continue
				}
				paths = append(paths, name)
				urlByPath[name] = u
				report.Files = append(report.Files, webaudit.FileResult{URL: u, Kind: kind, Bytes: len(data)})
			}
			report.URLsScanned = len(paths)

			cfg := rt.cfg
			cfg.Paths = paths
			cfg.LLM.Mode = "off"
			cfg.LayerB = false
			eng, err := engine.New(cfg)
			if err != nil {
				return err
			}
			defer eng.Close()
			res, err := eng.Scan(ctx)
			if err != nil {
				return err
			}
			// Remap finding paths to URLs for display
			for i := range res.Findings {
				if u, ok := urlByPath[res.Findings[i].Span.File]; ok {
					res.Findings[i].Span.File = u
				}
			}

			wantJSON := false
			if f := cmd.Root().PersistentFlags().Lookup("json"); f != nil && f.Changed {
				wantJSON, _ = cmd.Root().PersistentFlags().GetBool("json")
			}
			if format, _ := cmd.Root().PersistentFlags().GetString("format"); format == "json" {
				wantJSON = true
			}
			if wantJSON {
				out := map[string]any{
					"sitemap":        report.Sitemap,
					"base":           report.Base,
					"urls_collected": report.URLsCollected,
					"urls_scanned":   report.URLsScanned,
					"urls_failed":    report.Failures,
					"files":          report.Files,
					"findings":       res.Findings,
					"score":          res.Score,
					"disclaimer":     "Remote audit: metadata/c2patool optional tools not invoked.",
				}
				enc := json.NewEncoder(rt.opts.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			fmt.Fprintf(rt.opts.Stdout, "Sitemap: %s\nURLs collected: %d\nURLs scanned: %d\nURLs failed: %d\nFindings: %d\n",
				report.Sitemap, report.URLsCollected, report.URLsScanned, len(report.Failures), len(res.Findings))
			for _, f := range report.Failures {
				fmt.Fprintf(rt.opts.Stdout, "  [error] %s: %s\n", f.URL, f.Error)
			}
			if err := engine.Report(rt.opts.Stdout, "table", res, rt.opts.Version); err != nil {
				return err
			}
			if res.ShouldFail(rt.cfg.FailOn) || reportHasActionable(res) {
				return exitError{code: ExitFound, msg: ""}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&sitemap, "sitemap", "", "sitemap URL")
	cmd.Flags().StringVar(&base, "base", "", "base URL; discover sitemap automatically")
	cmd.Flags().IntVar(&maxPages, "max-pages", webaudit.DefaultMaxPages, "max URLs to fetch")
	cmd.Flags().DurationVar(&timeout, "timeout", webaudit.DefaultTimeout, "per-request timeout")
	cmd.Flags().IntVar(&maxBytes, "max-bytes", webaudit.DefaultMaxBytes, "max download size per URL")
	return cmd
}

func reportHasActionable(res *engine.Result) bool {
	for _, f := range res.Findings {
		if f.Confidence == domain.ConfidenceCertain || f.Confidence == domain.ConfidenceLikely {
			return true
		}
	}
	return false
}
