<h1 align="center">blotless/cli</h1>

<p align="center">
  <strong>CLI для blotless</strong><br>
  Inspect → clean следов AI. Zero CGO.
</p>

<p align="center">
  <b>Язык:</b> <a href="README.md">English</a> | Русский
</p>

---

**cli** — отдельный репозиторий. Детект в [engine](https://github.com/blotless/engine), трансформы в [ast](https://github.com/blotless/ast).

```bash
CGO_ENABLED=0 go install github.com/blotless/cli/cmd/blotless@latest
blotless inspect .
blotless clean . --write --layer-b
```

## WASM-плагин языка

Полный пример модуля: [ast/examples/wasm-rust](https://github.com/blotless/ast/tree/main/examples/wasm-rust).

```bash
blotless clean ./src --write --layer-b \
  --ast-wasm rust=./blotless_rust_transform.wasm
# .rs мапится автоматически

blotless clean . --write --layer-b \
  --ast-wasm zig=./zig_transform.wasm \
  --ast-ext .zig=zig
```

Документы этого репо: [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), [ROADMAP.md](ROADMAP.md), [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).  
Полное описание флагов и CI — в [README.md](README.md).
