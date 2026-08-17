# Security Policy

## Supported versions

Latest release of `github.com/blotless/cli` and `main`.

## Reporting

**Do not** file a public issue for vulnerabilities.

1. Email: **zikmanv@icloud.com**
2. Advisory: https://github.com/blotless/cli/security/advisories/new

Include `blotless version` output, impact, and reproduction.

**Initial response:** within 72 hours.

## Considerations

- Official binaries: `CGO_ENABLED=0`
- Network is opt-in (`--llm`, remote rewrite, `audit web`)
- `audit web` is SSRF-hardened; treat fetched content as untrusted
- `--ast-wasm` only loads **local** `.wasm` (no URLs)
- Layer B cannot certify vendor-detector failure
