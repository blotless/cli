# Architecture — blotless/cli

This repository is **only** the CLI. Libraries are published separately:

```
github.com/blotless/cli     ← you are here (Cobra, flags, reports)
        │
        ▼
github.com/blotless/engine  ← Scan / Clean / score
        │
        ▼
github.com/blotless/ast     ← Transform + WASM host
```

## Commands

| Command | Engine entry |
|---------|----------------|
| `inspect` / `scan` | `engine.Scan` (no writes, no Ollama) |
| `clean` | `engine.Clean` (`Write`, `LayerB`, `AstWASM`, `AstExt`) |
| `rewrite` | `engine/rewrite` single-file hook |
| `audit web` | `engine/webaudit` |
| `rules` / `version` | catalog / build metadata |

## Layer B flags

- `--layer-b` → `Config.LayerB`
- `--ast-wasm rust=./x.wasm` → `Config.AstWASM` → `ast.RegisterWASM` (`.rs` auto-mapped)
- `--ast-ext .zig=zig` → `Config.AstExt` → `ast.MapExt`
- `--llm` → optional paraphrase after AST transform

Plugin tutorial: [blotless/ast README](https://github.com/blotless/ast#add-a-language-via-wasm-no-go-required).

## Non-goals

No detectors in this repo. No raw exec of language tools.
