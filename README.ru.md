<h1 align="center">blotless/cli</h1>

<p align="center">
  <strong>CLI для blotless</strong><br>
  Сначала inspect, потом clean следов AI. Zero CGO.
</p>

<p align="center">
  <b>Язык:</b> <a href="README.md">English</a> | Русский
</p>

<p align="center">
  <a href="https://github.com/blotless/cli/actions/workflows/ci.yml"><img src="https://github.com/blotless/cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/blotless/cli"><img src="https://pkg.go.dev/badge/github.com/blotless/cli.svg" alt="Go Reference"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <a href="https://github.com/blotless/cli"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go" alt="Go Version"></a>
</p>

---

## Обзор

**cli** — командная строка экосистемы [blotless](https://github.com/blotless). Бинарник: `blotless` (`cmd/blotless`). Детекция и очистка — в [blotless/engine](https://github.com/blotless/engine).

### Возможности

| Категория | Что умеет |
|----------|-----------|
| **Команды** | `inspect` (`scan`), `clean`, `rewrite`, `audit web`, `rules`, `version` |
| **Inspect** | Все три слоя; не пишет; не запускает Ollama |
| **Clean** | По умолчанию dry-run; `--write` / `--in-place`; `--nfkc`; `--layer-b`; `--as` |
| **Rewrite** | Layer B: `print-prompt` (дефолт) / `ollama` / `openai` |
| **Отчёты** | table, JSON, YAML, SARIF |
| **Конфиг** | `.blotless.yaml` + `BLOTLESS_*` |
| **Сборка** | `CGO_ENABLED=0`, Go 1.26+; `version --json` |

---

## Установка

```bash
CGO_ENABLED=0 go install github.com/blotless/cli/cmd/blotless@latest
```

Из репозитория:

```bash
unset GOROOT
export GOTOOLCHAIN=local
CGO_ENABLED=0 go build -o blotless ./cmd/blotless
```

**Требования:** Go 1.26+, `CGO_ENABLED=0`.

---

## Быстрый старт

```bash
blotless inspect .
blotless inspect . --aggressive
blotless clean . --write --aggressive --nfkc
blotless rewrite draft.md --backend print-prompt --strength paraphrase
blotless clean . --layer-b
blotless clean . --write --llm=ollama --layer-b
blotless audit web --sitemap https://example.com/sitemap.xml --json
```

Layer B по умолчанию делает **агент** (скилл). Ollama — только по явному запросу.

---

## Экосистема

| Проект | Описание |
|--------|----------|
| [blotless/cli](https://github.com/blotless/cli) | **CLI (этот репозиторий)** |
| [blotless/engine](https://github.com/blotless/engine) | Детекция и очистка |
| [blotless/skills](https://github.com/blotless/skills) | Скилл агента → `~/.agent/skills/` |

---

## Лицензия

MIT — см. [LICENSE](LICENSE).
