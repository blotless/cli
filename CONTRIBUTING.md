# Contributing to blotless/cli

Thank you for contributing to the blotless command-line interface.

---

## Requirements

- **Go 1.26+**
- **`CGO_ENABLED=0`**
- **[Task](https://taskfile.dev)** (recommended)
- Depends on [blotless/engine](https://github.com/blotless/engine) (and transitively [blotless/ast](https://github.com/blotless/ast))

```bash
git clone https://github.com/blotless/cli
cd cli
unset GOROOT
export GOTOOLCHAIN=local
export CGO_ENABLED=0
task preflight   # or: go test ./...
```

---

## Workflow

```bash
git checkout -b feat/your-feature
gofmt -w .
go test ./...
git commit -m "feat(clean): add --ast-ext flag"
git push origin feat/your-feature
```

PR: https://github.com/blotless/cli/compare

### Checklist

- [ ] Tests pass (`go test ./...` / `task preflight`)
- [ ] Docs / README flags updated
- [ ] No `Co-authored-by:` trailers

Titles: `feat(inspect): …`, `fix(clean): …`, `docs: …`

Scopes: `inspect`, `clean`, `rewrite`, `audit`, `config`.

**Author:** `lkmavi <zikmanv@icloud.com>` unless agreed. Never `Co-authored-by:`.

Keep the public command surface stable: `inspect`, `clean`, `rewrite`, `audit web`, `rules`, `version`.

Cobra/Viper stay in this repo; detection logic belongs in **engine**, transforms in **ast**.

---

## Questions

- https://github.com/blotless/cli/issues
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

MIT: [LICENSE](LICENSE).
