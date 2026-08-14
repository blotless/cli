package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blotless/engine/rewrite"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (rt *runtime) rewriteCmd() *cobra.Command {
	var (
		backend, strength, output, model, endpoint, backup string
		lang, originalLang                                 string
		allowRemote, inPlace, stats, forceText, noLayerA   bool
		temperature                                        float64
		candidates                                         int
		timeout                                            time.Duration
	)
	cmd := &cobra.Command{
		Use:   "rewrite [file|-]",
		Short: "Layer B rewrite hook (statistical token watermarks)",
		Long: `Optional Layer B hook (WR rewrite_text.py).

Backends:
  print-prompt   emit prompt only (default; CI-safe, no model)
  ollama         POST to Ollama /api/chat
  openai         OpenAI-compatible /v1/chat/completions (alias: native)

Default backend is print-prompt — the agent (or you) runs the prompt.
After a model rewrite, Layer A scrub runs unless --no-layer-a-after.`,
		Example: `  blotless rewrite draft.md --backend print-prompt --strength paraphrase
  blotless rewrite draft.md -o draft.rewritten.md --backend ollama --model llama3.2
  blotless rewrite - --backend print-prompt < notes.txt`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "-"
			if len(args) > 0 {
				path = args[0]
			}
			raw, err := ReadTextInput(path, rt.opts.Stdin, forceText || rt.cfg.ForceText)
			if err != nil {
				return err
			}
			b, err := rewrite.NormalizeBackend(backend)
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("model") {
				if v := strings.TrimSpace(viper.GetString("llm-model")); v != "" {
					model = v
				}
			}
			if !cmd.Flags().Changed("endpoint") {
				if v := strings.TrimSpace(viper.GetString("llm-endpoint")); v != "" {
					endpoint = v
				}
			}
			if model == "" {
				model = strings.TrimSpace(os.Getenv("BLOTLESS_LLM_MODEL"))
			}
			apiKey := strings.TrimSpace(os.Getenv("BLOTLESS_REWRITE_API_KEY"))
			if apiKey == "" {
				apiKey = strings.TrimSpace(os.Getenv("BLOTLESS_LLM_API_KEY"))
			}

			res, err := rewrite.Run(cmd.Context(), string(raw), rewrite.Options{
				Backend:      b,
				Strength:     strength,
				Model:        model,
				Endpoint:     endpoint,
				APIKey:       apiKey,
				Timeout:      timeout,
				Temperature:  temperature,
				Candidates:   candidates,
				Lang:         lang,
				OriginalLang: originalLang,
				AllowRemote:  allowRemote,
				LayerAAfter:  !noLayerA && b != rewrite.BackendPrintPrompt,
				Aggressive:   rt.cfg.Aggressive,
			})
			if err != nil {
				return err
			}

			outPath := output
			if inPlace {
				if path == "" || path == "-" {
					return fmt.Errorf("--in-place requires a file path")
				}
				if _, err := BackupPath(path, backup); err != nil {
					return err
				}
				outPath = path
			} else if outPath == "" && path != "" && path != "-" && b != rewrite.BackendPrintPrompt {
				outPath = DerivedPath(path, ".rewritten")
			} else if outPath == "" {
				outPath = "-"
			}

			if err := WriteTextOutput(rt.opts.Stdout, outPath, []byte(res.Text)); err != nil {
				return err
			}
			if stats {
				enc := json.NewEncoder(rt.opts.Stderr)
				enc.SetIndent("", "  ")
				_ = enc.Encode(res.Info)
			} else {
				outChars := res.Info["output_chars"]
				if outChars == nil {
					outChars = len([]rune(res.Text))
				}
				fmt.Fprintf(rt.opts.Stderr, "backend=%v strength=%v mode=%v chars %v->%v\n",
					res.Info["backend"], res.Info["strength"], res.Info["mode"],
					res.Info["input_chars"], outChars)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&backend, "backend", rewrite.BackendPrintPrompt, "print-prompt|ollama|openai")
	cmd.Flags().StringVar(&strength, "strength", rewrite.StrengthParaphrase, "paraphrase|humanize|code|backtranslate|structural")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (default: stdout or *.rewritten.*)")
	cmd.Flags().BoolVar(&inPlace, "in-place", false, "overwrite input (writes backup first)")
	cmd.Flags().StringVar(&backup, "backup", ".bak", "backup suffix with --in-place")
	cmd.Flags().StringVar(&model, "model", "", "model name (required for ollama/openai)")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "base URL (default loopback ollama :11434 / openai :8080)")
	cmd.Flags().BoolVar(&allowRemote, "allow-remote", false, "allow non-loopback rewrite endpoints")
	cmd.Flags().Float64Var(&temperature, "temperature", 0.9, "sampling temperature")
	cmd.Flags().IntVar(&candidates, "candidates", 1, "number of rewrite candidates to score")
	cmd.Flags().StringVar(&lang, "lang", "French", "pivot language for backtranslate")
	cmd.Flags().StringVar(&originalLang, "original-lang", "English", "original language for backtranslate")
	cmd.Flags().DurationVar(&timeout, "timeout", 120*time.Second, "HTTP timeout")
	cmd.Flags().BoolVar(&noLayerA, "no-layer-a-after", false, "skip Layer A scrub on model output")
	cmd.Flags().BoolVar(&stats, "stats", false, "JSON stats on stderr")
	cmd.Flags().BoolVar(&forceText, "force-text", false, "rewrite even when input looks binary")
	return cmd
}
