<h1 align="center">blotless/cli</h1>

<p align="center">
  <strong>Command-line interface for blotless</strong><br>
  Inspect, then clean forensic AI traces. Zero CGO.
</p>

<p align="center">
  <b>Language:</b> English | <a href="README.ru.md">Русский</a>
</p>

<p align="center">
  <a href="https://github.com/blotless/cli/actions/workflows/ci.yml"><img src="https://github.com/blotless/cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/blotless/cli"><img src="https://pkg.go.dev/badge/github.com/blotless/cli.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/blotless/cli"><img src="https://goreportcard.com/badge/github.com/blotless/cli" alt="Go Report Card"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="https://github.com/blotless/cli/releases"><img src="https://img.shields.io/github/v/release/blotless/cli" alt="Latest Release"></a>
  <a href="https://github.com/blotless/cli"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go" alt="Go Version"></a>
</p>

---

## Overview

**cli** is the command-line interface of the [blotless](https://github.com/blotless) ecosystem. The binary is `blotless` (`cmd/blotless`). Detection and cleaning live in [blotless/engine](https://github.com/blotless/engine).

### Key Features

| Category | Capabilities |
|----------|--------------|
| **Commands** | `inspect` (`scan`), `clean`, `rewrite`, `audit web`, `rules`, `version` |
| **Inspect** | All three layers; never writes; never starts Ollama |
| **Clean** | Dry-run by default; `--write` / `--in-place`; `--nfkc`; `--layer-b`; `--as` |
| **Rewrite** | Layer B hook: `print-prompt` (default) / `ollama` / `openai`; strengths include backtranslate/structural |
| **Reports** | table, JSON, YAML, SARIF |
| **Config** | `.blotless.yaml` + `BLOTLESS_*` env |
| **Build** | `CGO_ENABLED=0`, Go 1.26+; `version --json` build metadata |

---

## Installation

```bash
CGO_ENABLED=0 go install github.com/blotless/cli/cmd/blotless@latest
```

From this repository:

```bash
unset GOROOT
export GOTOOLCHAIN=local
CGO_ENABLED=0 go build -o blotless ./cmd/blotless
```

**Requirements:** Go 1.26+, `CGO_ENABLED=0`.

> **Note:** On macOS and some Linux distros, CGO is enabled by default. Always set `CGO_ENABLED=0` when building blotless.

---

## Quick Start

```bash
blotless inspect .
blotless inspect . --aggressive
blotless clean . --write --aggressive --nfkc
blotless rewrite draft.md --backend print-prompt --strength paraphrase
blotless rewrite draft.md --backend ollama --model llama3.2 -o draft.rewritten.md
blotless clean . --layer-b                  # agent rewrites Layer B
blotless clean . --write --llm=ollama --layer-b
blotless audit web --sitemap https://example.com/sitemap.xml --json
```

`rewrite` is the single-file Layer B hook (WR `rewrite_text.py`). Tree LLM flags stay on **`clean`**.

---

## Layer B rewrite

| Flag | Default | Meaning |
|------|---------|---------|
| `--backend` | `print-prompt` | `print-prompt` \| `ollama` \| `openai` |
| `--strength` | `paraphrase` | `paraphrase` \| `humanize` \| `code` \| `backtranslate` \| `structural` |
| `-o` / `--output` | stdout / `*.rewritten.*` | output path |
| `--candidates` | `1` | N rewrites + lexical pick |
| `--allow-remote` | false | non-loopback endpoints |
| `--no-layer-a-after` | false | skip A scrub on model output |

---

## Ecosystem

| Project | Description |
|---------|-------------|
| [blotless/cli](https://github.com/blotless/cli) | **Command-line interface (this repo)** |
| [blotless/engine](https://github.com/blotless/engine) | Detection, cleaning, scoring, optional LLM (`internal/astgo`) |
| [blotless/skills](https://github.com/blotless/skills) | Agent skill — copy `remove-ai-marks` to `~/.agent/skills/` |

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>blotless</strong> — inspect first, then clean
</p>
