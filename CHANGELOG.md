# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-08-17

### Added

- `clean --layer-b` runs offline AST transform (Go/Python) without requiring `--llm`.
- `--ast-wasm lang=path.wasm` and `--ast-ext .ext=lang` for WASM language plugins.
- Inspect report: origin confidence and evidence.
- Contributor docs: CONTRIBUTING, SECURITY, ROADMAP, docs/ARCHITECTURE.

### Changed

- Depends on [engine v0.2.0](https://github.com/blotless/engine/releases/tag/v0.2.0).
- deps: cobra v1.9.1 → v1.10.2; viper v1.20.1 → v1.21.0.

## [0.1.0] - 2026-08-14

### Added

- Commands: inspect, clean, rewrite, audit web, rules, version.
