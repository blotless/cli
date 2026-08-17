package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/blotless/engine"
	"github.com/blotless/engine/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	ExitOK    = 0
	ExitFound = 1
	ExitErr   = 2
)

// Options inject IO and version for tests.
type Options struct {
	Version string
	Stdout  io.Writer
	Stderr  io.Writer
	Stdin   io.Reader
}

type runtime struct {
	opts Options
	cfg  engine.Config
}

// Execute runs the Cobra command tree and returns a process exit code.
func Execute(args []string, opts Options) int {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	if opts.Version == "" {
		opts.Version = Version
	}
	viper.Reset()
	rt := &runtime{opts: opts}
	root := rt.rootCmd()
	root.SetArgs(args)
	root.SetOut(opts.Stdout)
	root.SetErr(opts.Stderr)
	root.SetIn(opts.Stdin)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := root.ExecuteContext(ctx); err != nil {
		if code, ok := err.(exitError); ok {
			if code.msg != "" {
				fmt.Fprintln(opts.Stderr, err)
			}
			return code.code
		}
		fmt.Fprintln(opts.Stderr, err)
		return ExitErr
	}
	return ExitOK
}

type exitError struct {
	code int
	msg  string
}

func (e exitError) Error() string { return e.msg }

func (rt *runtime) rootCmd() *cobra.Command {
	var (
		configPath string
		verbose    bool
	)
	cmd := &cobra.Command{
		Use:     "blotless",
		Aliases: []string{"blot"},
		Short:   "Inspect then clean AI watermarks (Unicode, C2PA, statistical rewrite)",
		Long: `Inspect all three layers, then clean.

  A      invisible Unicode, stamps (verifiable)
  Files  C2PA / EXIF / XMP / doc props (verifiable)
  B      statistical token watermarks — no public detector;
         inspect reports which files are eligible for rewrite

inspect never writes files and never starts Ollama.
clean removes A+Files; Layer B uses the agent or: blotless rewrite / clean --llm=ollama.`,
		Example: `  blotless inspect .
  blotless inspect . --aggressive
  blotless clean . --write --aggressive --nfkc
  blotless rewrite draft.md --backend print-prompt
  blotless clean . --write --llm=ollama --layer-b`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := rt.loadConfig(configPath); err != nil {
				return err
			}
			rt.cfg.Aggressive = viper.GetBool("aggressive") || viper.GetBool("aggressive-homoglyphs") || viper.GetBool("strip-emoji-glue")
			rt.cfg.FailOn = domain.Confidence(viper.GetString("fail-on"))
			rt.cfg.Include = flagOrConfigStrings(cmd, "include")
			rt.cfg.Exclude = flagOrConfigStrings(cmd, "exclude")
			rt.cfg.DisabledRules = flagOrConfigStrings(cmd, "disable-rule")
			if rt.cfg.FailOn == "" {
				rt.cfg.FailOn = domain.ConfidenceNone
			}
			rt.cfg.Logger = newLogger(rt.opts.Stderr, verbose || viper.GetBool("verbose"))
			rt.cfg.ForceText = viper.GetBool("force-text")
			rt.cfg.LLM.Mode = "off"
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file (default .blotless.yaml)")
	cmd.PersistentFlags().String("format", "table", "output format: table|json|yaml|sarif")
	cmd.PersistentFlags().Bool("json", false, "alias for --format json")
	cmd.PersistentFlags().String("fail-on", "none", "exit 1 when findings meet confidence: none|certain|likely|any")
	cmd.PersistentFlags().Bool("aggressive", false, "Layer A: bidi in RTL, emoji glue, Go strings, Latin confusables")
	cmd.PersistentFlags().Bool("aggressive-homoglyphs", false, "alias of --aggressive")
	cmd.PersistentFlags().Bool("strip-emoji-glue", false, "alias of --aggressive")
	cmd.PersistentFlags().String("include", "", "comma-separated include globs")
	cmd.PersistentFlags().String("exclude", "", "exclude globs (CSV or YAML list); *_test.go matches any directory")
	cmd.PersistentFlags().String("disable-rule", "", "comma-separated rule IDs to skip")
	cmd.PersistentFlags().Bool("force-text", false, "scan unknown binaries as text")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "debug logs to stderr")
	_ = viper.BindPFlags(cmd.PersistentFlags())

	inspect := rt.inspectCmd()
	inspect.Use = "inspect [path ...]"
	inspect.Aliases = []string{"scan"}
	inspect.Short = "Look at layers A + Files + B eligibility (does not write)"
	cmd.AddCommand(inspect)
	cmd.AddCommand(rt.cleanCmd())
	cmd.AddCommand(rt.rewriteCmd())
	cmd.AddCommand(rt.auditCmd())
	cmd.AddCommand(rt.rulesCmd())
	cmd.AddCommand(rt.versionCmd())
	return cmd
}

func (rt *runtime) loadConfig(explicit string) error {
	viper.SetEnvPrefix("BLOTLESS")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	if explicit != "" {
		viper.SetConfigFile(explicit)
		return viper.ReadInConfig()
	}
	viper.SetConfigName(".blotless")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return err
	}
	return nil
}

func (rt *runtime) inspectCmd() *cobra.Command {
	var asKind string
	cmd := &cobra.Command{
		Use:   "inspect [path ...]",
		Short: "Look at layers A + Files + B eligibility (does not write)",
		Long: `Report all three layers without changing files.

  A      Unicode / stamps found
  Files  C2PA / metadata found
  B      files eligible for statistical rewrite (not detectable without a vendor key)

Ollama is not started. Use rewrite or clean --llm=ollama for Layer B.`,
		Example: `  blotless inspect .
  blotless inspect . --aggressive
  blotless inspect draft.md --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt.cfg.Paths = args
			if len(rt.cfg.Paths) == 0 {
				rt.cfg.Paths = []string{"."}
			}
			kind, err := ForceKind(asKind)
			if err != nil {
				return err
			}
			rt.cfg.ForceKind = kind
			rt.cfg.LLM.Mode = "off"
			rt.cfg.LayerB = false
			format := outputFormat(cmd)
			rt.attachProgress(format)
			eng, err := engine.New(rt.cfg)
			if err != nil {
				return err
			}
			defer eng.Close()
			res, err := eng.Scan(cmd.Context())
			if err != nil {
				return err
			}
			if err := engine.Report(rt.opts.Stdout, format, res, rt.opts.Version); err != nil {
				return err
			}
			if res.ShouldFail(rt.cfg.FailOn) {
				return exitError{code: ExitFound, msg: ""}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&asKind, "as", "auto", "force kind: auto|text|image|container")
	return cmd
}

func (rt *runtime) cleanCmd() *cobra.Command {
	var write, inPlace, nfkc, layerB bool
	var backup, strength, llm, llmModel, llmEndpoint, asKind string
	var llmTimeout time.Duration
	var astWASM, astExt []string
	cmd := &cobra.Command{
		Use:   "clean [path ...]",
		Short: "Remove Layer A + Files; optional Layer B via AST transform / --llm",
		Long: `Strip Unicode/stamps and file metadata. Default is dry-run.

Layer B (statistical) rewrite:
  --layer-b      AST transform for Go/Python (offline) + optional LLM
  (no --llm)     AST transform still applies; prose rewrite in the agent
  --llm=ollama   local Ollama (tree walk)
  --llm=native   OpenAI-compatible llama.cpp
  --strength     paraphrase | humanize | code | backtranslate | structural
  --ast-wasm     lang=path.wasm (capability-sandboxed plugin)
  --ast-ext      .rs=rust (map extra extensions; .rs is auto-mapped for rust)

Single-file Layer B hook: blotless rewrite (print-prompt default).`,
		Example: `  blotless clean .
  blotless clean . --write --aggressive --nfkc
  blotless clean . --write --layer-b
  blotless clean . --write --llm=ollama --layer-b --strength paraphrase
  blotless clean . --write --layer-b --ast-wasm rust=./rust_transform.wasm
  blotless clean . --write --layer-b --ast-wasm zig=./zig_transform.wasm --ast-ext .zig=zig`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt.cfg.Paths = args
			if len(rt.cfg.Paths) == 0 {
				rt.cfg.Paths = []string{"."}
			}
			kind, err := ForceKind(asKind)
			if err != nil {
				return err
			}
			rt.cfg.ForceKind = kind
			rt.cfg.Write = write || inPlace
			rt.cfg.Backup = backup
			rt.cfg.NFKC = nfkc
			rt.cfg.LayerB = layerB
			strength = strings.ToLower(strings.TrimSpace(strength))
			if strength == "" {
				strength = "paraphrase"
			}
			switch strength {
			case "paraphrase", "humanize", "code", "backtranslate", "structural":
			default:
				return fmt.Errorf("--strength must be paraphrase, humanize, code, backtranslate, or structural")
			}
			rt.cfg.LayerBStrength = strength
			rt.cfg.AstWASM = map[string]string{}
			for _, pair := range astWASM {
				lang, path, ok := strings.Cut(pair, "=")
				if !ok || strings.TrimSpace(lang) == "" || strings.TrimSpace(path) == "" {
					return fmt.Errorf("--ast-wasm wants lang=path.wasm, got %q", pair)
				}
				rt.cfg.AstWASM[strings.TrimSpace(lang)] = strings.TrimSpace(path)
			}
			rt.cfg.AstExt = map[string]string{}
			for _, pair := range astExt {
				ext, lang, ok := strings.Cut(pair, "=")
				if !ok || strings.TrimSpace(ext) == "" || strings.TrimSpace(lang) == "" {
					return fmt.Errorf("--ast-ext wants .ext=lang, got %q", pair)
				}
				rt.cfg.AstExt[strings.TrimSpace(ext)] = strings.TrimSpace(lang)
			}
			if !cmd.Flags().Changed("llm") {
				if v := strings.TrimSpace(viper.GetString("llm")); v != "" {
					llm = v
				}
			}
			if !cmd.Flags().Changed("llm-model") {
				llmModel = viper.GetString("llm-model")
			}
			if !cmd.Flags().Changed("llm-endpoint") {
				llmEndpoint = viper.GetString("llm-endpoint")
			}
			if llm == "" {
				llm = "off"
			}
			rt.cfg.LLM = engine.LLMConfig{
				Mode:     llm,
				Model:    llmModel,
				Endpoint: llmEndpoint,
				Timeout:  llmTimeout,
			}
			format := outputFormat(cmd)
			rt.attachProgress(format)
			eng, err := engine.New(rt.cfg)
			if err != nil {
				return err
			}
			defer eng.Close()
			res, err := eng.Clean(cmd.Context())
			if err != nil {
				return err
			}
			if err := engine.Report(rt.opts.Stdout, format, res, rt.opts.Version); err != nil {
				return err
			}
			if res.ShouldFail(rt.cfg.FailOn) {
				return exitError{code: ExitFound, msg: ""}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&write, "write", false, "apply patches (default is dry-run)")
	cmd.Flags().BoolVar(&inPlace, "in-place", false, "alias for --write")
	cmd.Flags().BoolVar(&nfkc, "nfkc", false, "Layer A: NFKC-normalize text after strip")
	cmd.Flags().StringVar(&strength, "strength", "paraphrase", "Layer B rewrite: paraphrase|humanize|code|backtranslate|structural")
	cmd.Flags().StringVar(&backup, "backup", "", "backup suffix, e.g. .bak")
	cmd.Flags().BoolVar(&layerB, "layer-b", false, "Layer B even when A is clean (AST transform offline; LLM optional)")
	cmd.Flags().StringVar(&llm, "llm", "off", "Layer B backend: off (agent) | ollama | native")
	cmd.Flags().StringVar(&llmModel, "llm-model", "", "model name (default qwen2.5-coder:3b)")
	cmd.Flags().StringVar(&llmEndpoint, "llm-endpoint", "", "Ollama URL; empty reuses :11434 or starts ollama")
	cmd.Flags().DurationVar(&llmTimeout, "llm-timeout", 30*time.Second, "LLM request timeout")
	cmd.Flags().StringArrayVar(&astWASM, "ast-wasm", nil, "WASM transform plugin lang=path.wasm (repeatable)")
	cmd.Flags().StringArrayVar(&astExt, "ast-ext", nil, "map extra extension to lang, e.g. .zig=zig (repeatable)")
	cmd.Flags().StringVar(&asKind, "as", "auto", "force kind: auto|text|image|container")
	return cmd
}

func outputFormat(cmd *cobra.Command) string {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("json"); f != nil && f.Changed {
			if jsonFlag, _ := cmd.Root().PersistentFlags().GetBool("json"); jsonFlag {
				return "json"
			}
		}
	}
	format := viper.GetString("format")
	if format == "" && cmd != nil {
		format, _ = cmd.Flags().GetString("format")
	}
	if format == "" {
		return "table"
	}
	return format
}

func (rt *runtime) attachProgress(format string) {
	ui := newLiveUI(rt.opts.Stdout, rt.opts.Stderr)
	if strings.ToLower(format) != "table" {
		rt.cfg.Progress = barOnly{ui}
		return
	}
	rt.cfg.Progress = ui
}

type barOnly struct{ *liveUI }

func (barOnly) Finding(domain.Finding) {}

func (rt *runtime) rulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "List or explain detection rules",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "ls",
		Short: "List rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := engine.Rules()
			if err != nil {
				return err
			}
			format := viper.GetString("format")
			if format == "json" {
				enc := json.NewEncoder(rt.opts.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(list)
			}
			if format == "yaml" || format == "yml" {
				enc := yaml.NewEncoder(rt.opts.Stdout)
				enc.SetIndent(2)
				defer enc.Close()
				return enc.Encode(list)
			}
			for _, r := range list {
				fmt.Fprintf(rt.opts.Stdout, "%s\t%s\t%s\t%s\n", r.ID, r.Family, r.Confidence, r.Description)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "explain RULE_ID",
		Short: "Show one rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := engine.Explain(args[0])
			if err != nil {
				return err
			}
			format := viper.GetString("format")
			if format == "yaml" || format == "yml" {
				enc := yaml.NewEncoder(rt.opts.Stdout)
				enc.SetIndent(2)
				defer enc.Close()
				return enc.Encode(r)
			}
			enc := json.NewEncoder(rt.opts.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(r)
		},
	})
	return cmd
}

func (rt *runtime) versionCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and build metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			bi := CurrentBuildInfo(rt.opts.Version)
			wantJSON := asJSON
			if f := cmd.Root().PersistentFlags().Lookup("json"); f != nil && f.Changed {
				wantJSON, _ = cmd.Root().PersistentFlags().GetBool("json")
			}
			_, err := fmt.Fprint(rt.opts.Stdout, formatBuildInfo(bi, wantJSON))
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit JSON build info")
	return cmd
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func flagOrConfigStrings(cmd *cobra.Command, key string) []string {
	if f := cmd.Root().PersistentFlags().Lookup(key); f != nil && f.Changed {
		return splitCSV(f.Value.String())
	}
	return parseViperStrings(key)
}

func parseViperStrings(key string) []string {
	raw := viper.Get(key)
	switch v := raw.(type) {
	case []string:
		return cleanStrings(v)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s := strings.TrimSpace(fmt.Sprint(item))
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		return splitCSV(v)
	}
	ss := viper.GetStringSlice(key)
	if len(ss) == 1 {
		return splitCSV(ss[0])
	}
	return cleanStrings(ss)
}

func cleanStrings(in []string) []string {
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
