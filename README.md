<p align="center">
  <strong>Command-line interface for blotless</strong><br>
  Inspect, then clean forensic AI traces. Zero CGO.
</p>

<h1 align="center">blotless/cli</h1>

<p align="center">
  Unicode Layer A · statistical Layer B · C2PA / files · AST transform · origin heuristics
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

**cli** is the command-line interface of the [blotless](https://github.com/blotless) ecosystem. The binary is `blotless` (`cmd/blotless`). Detection and cleaning live in [blotless/engine](https://github.com/blotless/engine); language transforms in [blotless/ast](https://github.com/blotless/ast). These are **separate repositories**.

### Key Features

| Category | Capabilities |
|----------|--------------|
| **Commands** | `inspect` (`scan`), `clean`, `rewrite`, `audit web`, `rules`, `version` |
| **Inspect** | Layers A / Files / B eligibility; origin confidence + evidence; never writes; never starts Ollama |
| **Clean** | Dry-run by default; `--write`; `--nfkc`; `--layer-b` (AST transform offline); `--ast-wasm`; `--llm` |
| **Rewrite** | Single-file Layer B hook: `print-prompt` / `ollama` / `openai` |
| **Reports** | table, JSON, YAML, SARIF |
| **Config** | `.blotless.yaml` + `BLOTLESS_*` |
| **Build** | `CGO_ENABLED=0`, Go 1.26+; `task preflight` |

---

## Installation

```bash
CGO_ENABLED=0 go install github.com/blotless/cli/cmd/blotless@latest
```

From a pinned release: [Releases](https://github.com/blotless/cli/releases) (preferred for corporate fleets).

From this repository:

```bash
git clone https://github.com/blotless/cli
cd cli
unset GOROOT
export GOTOOLCHAIN=local
CGO_ENABLED=0 go build -o blotless ./cmd/blotless
```

**Requirements:** Go 1.26+, `CGO_ENABLED=0`.

> On macOS and some Linux distros CGO is on by default. Always set `CGO_ENABLED=0`.

---

## Quick Start

```bash
blotless inspect .
blotless inspect . --aggressive --format json
blotless clean . --write --aggressive --nfkc
blotless clean . --write --layer-b
blotless clean . --write --llm=ollama --layer-b --strength paraphrase
blotless rewrite draft.md --backend print-prompt --strength paraphrase
blotless audit web --sitemap https://example.com/sitemap.xml --json
```

`scan` is an alias of `inspect`.

---

## Layer B

| Mode | Flag | Behavior |
|------|------|----------|
| AST transform | `--layer-b` | Offline Go/Python (and WASM plugins) via [blotless/ast](https://github.com/blotless/ast) |
| LLM rewrite | `--llm=ollama\|native` | Optional paraphrase on prose/comments |
| Agent | no `--llm` | AST still runs; prose rewrite can be done by the agent |

### Install a WASM language plugin

Full module example (Rust crate you can copy): [blotless/ast examples/wasm-rust](https://github.com/blotless/ast/tree/main/examples/wasm-rust). ABI: [ABI.md](https://github.com/blotless/ast/blob/main/ABI.md).

```bash
# 1. Build plugin (once)
rustup target add wasm32-unknown-unknown
cargo build --release --target wasm32-unknown-unknown --manifest-path examples/wasm-rust/Cargo.toml

# 2. Connect — rust id auto-maps .rs
blotless clean ./src --write --layer-b \
  --ast-wasm rust=./blotless_rust_transform.wasm

# 3. Other languages: map the extension
blotless clean . --write --layer-b \
  --ast-wasm zig=/opt/blotless/zig_transform.wasm \
  --ast-ext .zig=zig
```

Paths must be **local files**. Guests have no FS/network/env (wazero capability host).

---

## Flags

**Global**

| Flag | Meaning |
|------|---------|
| `--config` | YAML (default `.blotless.yaml`) |
| `--format table\|json\|yaml\|sarif` | Report format |
| `--json` | Alias of `--format json` |
| `--fail-on none\|certain\|likely\|any` | Exit 1 on findings |
| `--aggressive` | Stronger Layer A |
| `--include` / `--exclude` | Globs |
| `--disable-rule` | Skip rule ID |
| `-v` / `--verbose` | Debug on stderr |

**`clean`**

| Flag | Meaning |
|------|---------|
| `--write` / `--in-place` | Write patches (otherwise dry-run) |
| `--backup .bak` | Copy before writing |
| `--nfkc` | NFKC after strip |
| `--layer-b` | AST transform even if Layer A is clean |
| `--llm off\|ollama\|native` | Optional paraphrase backend |
| `--strength` | paraphrase \| humanize \| code \| backtranslate \| structural |
| `--ast-wasm lang=path.wasm` | WASM plugin (repeatable) |
| `--ast-ext .zig=zig` | Extra extension map (repeatable) |

Env prefix: `BLOTLESS_` (e.g. `BLOTLESS_LLM=ollama`).

---

## CI gate

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: "1.26.x"
- name: Install blotless
  run: CGO_ENABLED=0 go install github.com/blotless/cli/cmd/blotless@latest
- name: Gate on certain findings
  run: blotless inspect . --fail-on certain --format sarif > blotless.sarif
```

---

## Architecture

Thin Cobra wrapper around `engine.Scan` / `engine.Clean`. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## Ecosystem

| Project | Description |
|---------|-------------|
| [blotless/cli](https://github.com/blotless/cli) | **This repo** |
| [blotless/engine](https://github.com/blotless/engine) | Detection / clean library |
| [blotless/ast](https://github.com/blotless/ast) | AST transform + WASM host |
| [blotless/skills](https://github.com/blotless/skills) | Agent skill pack |

---

## Disclaimer

Layer B cannot certify vendor-detector failure. Origin / `likely agent` is heuristic. Do not claim results as human-authored.

---

## Community

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [SECURITY.md](SECURITY.md)
- [ROADMAP.md](ROADMAP.md)

---

## License

MIT License — see [LICENSE](LICENSE).

---

<p align="center"><strong>blotless</strong> — inspect first, then clean</p>
