# AGENTS.md

> `github.com/house/goscribe` — Go 1.26 CLI tool that reads source code and generates AI-powered documentation. Supports 16 providers via a registry pattern.

## Build & Test

```bash
go build ./cmd/goscribe                          # build binary
go test ./...                                    # unit tests
go test -race -cover ./...                       # with race + coverage
go test -tags=integration ./test/integration/... # integration tests (separate tag)
golangci-lint run ./...                          # lint (v2 config)
go vet ./...                                     # static analysis
gosec ./...                                      # security scan
```

## Architecture

| Package | Role | Boundary |
|---------|------|----------|
| `cmd/` | Cobra commands only. No business logic. | CLI surface |
| `internal/ai/` | Provider interface, registry, 16 implementations, retry | Private |
| `internal/docs/` | Generation engine, updater, state tracking | Private |
| `internal/config/` | Config loading (viper) | Private |
| `internal/ci/` | Structured output formatting (json/markdown/github/text) | Private |
| `internal/git/` | Git operations wrapper | Private |
| `internal/version/` | Version management + Scooby-Doo quotes | Private |
| `pkg/goscribe/` | Public library API. Wraps internals without exposing `internal/*` types | Public |
| `pkg/providers/` | Provider config persistence | Public |
| `plugin/opencode/` | Self-contained IDE plugin with event system + file watcher | Plugin |

Key patterns:
- **Registry pattern** for providers: `RegisterProvider(name, factory)` in `init()`. See `internal/ai/provider.go`.
- **Functional options** for `pkg/goscribe.NewClient`: `WithProvider`, `WithModel`, `WithOutputDir`, `WithTimeout`, `WithRetry`, `WithWorkers`, `WithCacheDir`, etc.
- **RetryProvider** wraps any Provider with exponential backoff + 20% jitter, 30s max. Non-retryable errors (auth, invalid model, context length) skip retries.
- **Git state** tracked in `.goscribe-state` (JSON: `{"last_commit":"hash"}`). `generate` saves hash; `update` diffs against HEAD.
- **Content-addressed cache** via `WithCacheDir`. SHA-256 keyed.

## Key Conventions

**Config priority:** flags > `GOSCRIBE_*` env vars > `~/.goscribe.yaml` > defaults.

**Exit codes:**
- `0` — success
- `1` — error (generation/update failed)
- `2` — update succeeded but no changes detected

`ExitCodeError` in `cmd/root.go` wraps exit codes for CI mode signaling.

**Version:** hardcoded default is `"0.0.1"`. Release builds inject via ldflags: `-X github.com/house/goscribe/internal/version.Version={{.Version}}`.

**TUI wizard quirk:** Running `goscribe` with no subcommand launches a Bubble Tea interactive wizard (`cmd/wizard/`). `--ci` or `--help` bypasses it. If no TTY, falls back to help text.

**CI mode:** `--ci` enables structured output. `--output-format` values: `text`, `json`, `markdown`, `github`.

## Always / Ask / Never

**Always:**
- Keep business logic in `internal/`. Keep `cmd/` Cobra-only.
- Add doc comments to exported types/functions (revive `exported` rule).
- Add package comments (revive `package-comments` rule).
- Use `filepath.Clean()` + prefix validation before file reads/writes.
- Use `crypto/rand` for jitter, not `math/rand`.
- Set file perms to `0750` (dirs) / `0600` (files) in production code.

**Ask:**
- Before adding new top-level packages.
- Before changing provider interface signatures.
- Before modifying CI job structure.

**Never:**
- Expose `internal/*` types in `pkg/goscribe/` exported signatures.
- Hardcode provider credentials.
- Skip `go vet` or `golangci-lint` before committing.

## Testing Conventions

**CLI-first QA:** Prove the command-line workflow first. For any user-facing change:
1. Test `goscribe --help` and command-specific help.
2. Test happy-path via built binary (`./goscribe ...`).
3. Test one invalid-input case for error message + exit behavior.
4. Test config/flag/env precedence for changed options.
5. Test `--ci` structured output if automation behavior changed.

**Integration tests:** Use `-tags=integration` to avoid binary rebuilds during unit test runs.

**gosec suppressions** (actual, from `#nosec` annotations):
- `G104` — viper `BindPFlag` calls (non-critical config binding)
- `G204` — `exec.Command("git", ...)` (hardcoded binary, no user injection)
- `G304` — file reads after `filepath.Clean()` + prefix traversal validation
- `G301` — test utility temp dirs (permissive modes)
- `G306` — test utility temp files (permissive modes)

## Adding a Provider

1. Implement the `Provider` interface in `internal/ai/<name>.go`.
2. Add a factory function.
3. Call `RegisterProvider("<name>", <name>Factory)` in the `init()` of `internal/ai/provider.go`.
4. The registry already has 16 providers. Follow the existing pattern.

## Build Flags

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/goscribe
```

Cross-compile targets (from CI matrix): `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`.

## CI/CD

5 jobs in `.github/workflows/ci.yml`:
1. **lint** — golangci-lint v2.1, 5m timeout
2. **test** — `go test -race -coverprofile=coverage.out -covermode=atomic ./...`
3. **security** — gosec
4. **integration** — `go test -tags=integration ./test/integration/...` (needs lint + test)
5. **build** — cross-compile matrix, needs all above

## Lint Config

`.golangci.yml` (v2 format). Enabled linters: errcheck, govet, ineffassign, staticcheck, unused, misspell, revive. Formatters: gofmt, goimports. Revive rules include `exported` and `package-comments`.
