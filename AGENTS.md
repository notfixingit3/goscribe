# AGENTS.md

> GoScribe v1.0 - AI-powered documentation generator for Go projects.

## Project Overview

- **Name:** goscribe
- **Language:** Go 1.26
- **Type:** CLI application, Go library, CI/CD tool, OpenCode plugin
- **Purpose:** Reads application source code and produces comprehensive user documentation using AI providers (OpenAI, Ollama). Designed for four usage modes: CLI, programmatic library, CI/CD pipeline integration, and OpenCode plugin.
- **Module:** `github.com/house/goscribe`
- **Status:** Production-ready. CI/CD, cross-platform builds, retries, timeouts all in place.

## Quick Start

### Build
```bash
go build ./cmd/goscribe
```

### Run
```bash
./goscribe generate [path]    # Generate docs from source code
./goscribe update [path]      # Update docs based on git changes
./goscribe provider add ...   # Configure AI providers
./goscribe version            # Show version
./goscribe bump               # Bump patch version + commit with Scooby-Doo quote
./goscribe tag                # Tag current commit with version
```

## Architecture

### Multi-Mode Diagram

```
                         GoScribe
                            |
         +------------------+------------------+
         |                  |                  |
    Source Code    AI Provider Interface    Git State
    (.go files)    (OpenAI / Ollama)        (.goscribe-state)
         |                  |                  |
         +------------------+------------------+
                            |
              +-------------+-------------+
              |             |             |
          CLI Mode     Library Mode    Plugin Mode
          (cobra)      (pkg/goscribe)  (plugin/opencode)
              |             |             |
              +------+------+------+------+
                     |             |
                 CI/CD Mode    Docker Mode
                 (--ci flag)   (multi-stage)
```

All four modes share the same provider interface and git state tracking. The library package (`pkg/goscribe`) wraps internal packages without exposing them. The CLI depends on internal packages directly. The plugin uses its own event system and file watcher.

### Project Structure
```
goscribe/
├── cmd/
│   ├── goscribe/
│   │   └── main.go          # Entry point (minimal, <50 lines)
│   ├── root.go              # Root command + config initialization
│   ├── generate.go          # Generate documentation command
│   ├── update.go            # Update documentation command
│   ├── provider.go          # Provider management commands
│   └── version.go           # Version + bump/tag commands
├── internal/
│   ├── ai/
│   │   ├── provider.go      # Provider interface + factory (includes retry wrapping)
│   │   ├── openai.go        # OpenAI client with timeout + enhanced error messages
│   │   ├── ollama.go        # Ollama client implementation
│   │   ├── retry.go         # Exponential backoff retry with jitter
│   │   └── retry_test.go    # Retry logic tests
│   ├── ci/
│   │   └── output.go        # Structured output formatting (json/markdown/github/text)
│   ├── config/
│   │   └── config.go        # Configuration loading/saving (timeout, retries, backoff, CI flags)
│   ├── docs/
│   │   ├── generator.go     # Documentation generation engine
│   │   ├── updater.go       # Incremental doc updates
│   │   └── state.go         # JSON-based git commit state tracking
│   ├── git/
│   │   └── git.go           # Git operations wrapper
│   └── version/
│       ├── version.go       # Version management + Scooby-Doo quotes
│       └── version_test.go  # Version tests
├── pkg/
│   ├── goscribe/
│   │   ├── client.go        # Public Client + NewClient constructor
│   │   ├── options.go       # Functional options (WithProvider, WithModel, etc.)
│   │   ├── generate.go      # Generate method + GenerateOptions/GenerateResult
│   │   ├── update.go        # Update method + UpdateOptions/UpdateResult
│   │   ├── errors.go        # Sentinel errors + Error struct (errors.Is compatible)
│   │   └── version.go       # Library version constant (currently "1.0.0")
│   └── providers/
│       └── providers.go     # Provider configuration persistence
├── plugin/
│   └── opencode/
│       ├── client.go        # Plugin client: Init, Run, Shutdown, RegisterHandler
│       ├── config.go        # PluginConfig, LoadConfig, DefaultConfig, Save
│       ├── types.go         # EventType, PluginState, Event, Handler, GenerationRequest/Response
│       └── watcher.go       # FileWatcher with debounce, .gitignore support
├── .github/
│   ├── actions/goscribe/
│   │   └── action.yml       # Composite action for CI workflows
│   └── workflows/
│       ├── ci.yml            # Lint, test, security, build
│       └── test-action.yml   # Action integration test
├── Dockerfile               # Multi-stage build (alpine, non-root user)
├── .golangci.yml            # golangci-lint configuration (v2 format)
└── ...
```

### Key Design Decisions

- **cmd/** contains Cobra commands only. No business logic.
- **internal/** contains all implementation details (compiler-enforced privacy).
- **pkg/goscribe/** is the stable public API for library usage. Wraps internals without exposing `internal/*` types in exported signatures.
- **pkg/providers/** handles provider configuration persistence (separate concern from the library API).
- **plugin/opencode/** is self-contained with its own config, event system, and file watcher. Uses the `pkg/goscribe` public API for document generation. Not on internal packages directly.
- **internal/ci/** handles structured output formatting. The CLI delegates to it when `--ci` is set. Library users get structured Go types instead.
- Configuration stored in `~/.goscribe.yaml`. Also looks for `./.goscribe.yaml` (local override).
- Git state tracked in `.goscribe-state` file (JSON format: `{"last_commit": "abc123"}`).
- RetryProvider wraps any Provider with exponential backoff + jitter. Factory in `provider.go` applies it when `retries > 0`.
- OpenAI client enhances errors with actionable suggestions (auth, rate limit, quota, model not found, context length, network).

## Dependencies

| Library | Purpose | Version |
|---------|---------|---------|
| cobra | CLI framework | v1.10.2 |
| viper | Configuration management | v1.21.0 |
| go-git/v5 | Git operations | v5.19.1 |
| openai-go | OpenAI API client | v1.12.0 |
| ollama/api | Ollama API client | v0.24.0 |

## Commands

### Generate Documentation
```bash
./goscribe generate                    # current directory
./goscribe generate /path/to/project   # specific project
./goscribe generate -f                 # force, overwrite existing docs
```

### Update Documentation (Git-aware)
```bash
./goscribe update
./goscribe update /path/to/project
```
Requires a git repository and a `.goscribe-state` file from a previous `generate` run.

### Provider Management
```bash
./goscribe provider add ollama --url http://localhost:11434 --model llama2
./goscribe provider add openai --key sk-... --model gpt-4
./goscribe provider list
./goscribe provider remove ollama
```

### Version Management
```bash
./goscribe version    # Print current version
./goscribe bump       # Bump patch (0.0.x), stage all, commit with Scooby-Doo quote
./goscribe tag        # Tag HEAD with current version
```

## Configuration

Config file: `~/.goscribe.yaml` (or `./.goscribe.yaml`, or pass `--config /path`).

```yaml
provider: ollama
model: llama2
output: docs
verbose: false
timeout: 5m
retries: 3
retry_backoff: 2s
providers:
  - name: ollama
    url: http://localhost:11434
    model: llama2
    default: true
  - name: openai
    api_key: sk-...
    model: gpt-4
```

Environment variables (prefix `GOSCRIBE_`):
- `GOSCRIBE_PROVIDER`, `GOSCRIBE_MODEL`, `GOSCRIBE_OUTPUT`, `GOSCRIBE_VERBOSE`
- `GOSCRIBE_TIMEOUT`, `GOSCRIBE_RETRIES`, `GOSCRIBE_RETRY_BACKOFF`

Priority: command flags > env vars > config file > defaults.

### Timeout and Retry Defaults

| Setting | Default | Config Key | Flag |
|---------|---------|------------|------|
| AI call timeout | 5m | `timeout` | `--timeout` |
| Retry attempts | 3 | `retries` | `--retries` |
| Initial backoff | 2s | `retry_backoff` | `--retry-backoff` |

Retry uses exponential backoff with 20% jitter, capped at 30s max wait. Non-retryable errors (auth, invalid model, context length exceeded) skip retries entirely.

## CI/CD

### GitHub Actions (`.github/workflows/ci.yml`)

Runs on push to `main` and PRs against `main`. Four jobs:

1. **Lint** - `golangci/golangci-lint-action@v7` (v2.1) with 5m timeout. Config in `.golangci.yml`.
2. **Test** - `go test -race -coverprofile=coverage.out -covermode=atomic ./...`, uploads coverage artifact (7-day retention)
3. **Security** - `gosec ./...` via `securego/gosec` action (runs in parallel with lint/test)
4. **Build** - Cross-compile matrix (only after lint+test+security pass):
   - linux/amd64, linux/arm64
   - darwin/amd64, darwin/arm64
   - windows/amd64, windows/arm64
   - Built with `CGO_ENABLED=0` and `-ldflags="-s -w"`

### GoReleaser (`.goreleaser.yaml`)

Configured for release builds. Key details:
- Runs `go mod tidy` as a pre-hook
- Builds with `-ldflags "-s -w -X github.com/house/goscribe/internal/version.Version={{.Version}}"` to inject the version at build time
- Outputs tar.gz (zip for Windows)
- Changelog excludes `docs:`, `test:`, `ci:` prefixed commits

## Release Process

1. `./goscribe bump` to bump patch version and commit
2. `./goscribe tag` to tag the commit
3. `git push --follow-tags` to push commit + tag
4. GoReleaser picks up the tag and builds cross-platform binaries
5. Version is baked into the binary via ldflags at release time. During development, `version.Version` stays at the hardcoded default (`0.0.1`).

## Git Integration

- **State file:** `.goscribe-state` in the source directory, JSON format (`{"last_commit": "hash"}`). Backwards compatible with old plain-text format (the loader falls back to reading raw content as a string).
- **First run:** `generate` saves current commit hash.
- **Updates:** `update` compares saved hash to HEAD, processes only changed files, then saves the new hash.

## Testing

### Current QA Focus: CLI First

For near-term testing and product validation, treat the CLI as the primary user surface. Library, CI/CD, Docker, and OpenCode plugin support must remain compatible, but new feature QA should prove the command-line workflow first.

CLI-first manual QA should cover:
- `goscribe --help` plus command-specific help for changed commands
- Happy-path command execution through the built binary (`./goscribe ...`)
- One invalid-input/error case with a useful message and exit behavior
- Config, flag, and `GOSCRIBE_*` environment variable precedence for changed options
- `--ci` structured output when a CLI change affects automation behavior

Prefer integration tests that execute the compiled CLI for user-facing behavior. Use direct package tests for internal logic, edge cases, and fast feedback, but do not consider a user-facing change complete until the CLI surface has been exercised.

```bash
go test ./...                           # run unit tests
go test -race -cover ./...              # with race detector and coverage
go test -tags=integration ./test/integration  # run integration tests
golangci-lint run ./...                 # run linters (requires golangci-lint v2)
go vet ./...                            # basic static analysis
gosec ./...                             # security scanning
```

CI runs tests with `-race -coverprofile=coverage.out -covermode=atomic`.
Integration tests run separately with `-tags=integration` to avoid binary rebuilds during unit test runs.

### Linting

The project uses golangci-lint v2 (config in `.golangci.yml`). Enabled linters:
errcheck, govet, ineffassign, staticcheck, unused, misspell, revive.
Formatters: gofmt, goimports.

Run locally:
```bash
golangci-lint run ./...                 # lint all
golangci-lint run --fix ./...           # auto-fix where possible
```

All exported types and functions must have doc comments (revive `exported` rule).
All packages must have package comments (revive `package-comments` rule).

### Security Scanning

The project uses `gosec` for static security analysis. Run locally with `gosec ./...`.

Known suppressions (via `#nosec` annotations):
- **G104** (errors unhandled): viper `BindPFlag` calls — non-critical config binding, standard cobra/viper pattern
- **G204** (subprocess): `exec.Command("git", ...)` — hardcoded binary name, no user-controlled command injection
- **G304/G703** (path traversal): file reads/writes after explicit `filepath.Clean()` + prefix-based traversal validation
- **G301/G306** (permissions): test utility files use permissive modes for temp directories

Retry jitter uses `crypto/rand` instead of `math/rand`. File permissions tightened to 0750 (dirs) and 0600 (files) in production code.

## Usage Modes

### 1. CLI Tool

Direct command-line usage. See the Commands section for full flag reference.

```bash
# Basic workflow
goscribe provider add ollama --url http://localhost:11434 --model llama2
goscribe generate              # generate docs for current directory
goscribe update                # incrementally update based on git diff
```

CI mode flags (see Mode 3 below for full CI/CD details):

```bash
goscribe generate --ci --output-format json
goscribe update --ci --output-format github
```

### 2. Go Library

Import `github.com/house/goscribe/pkg/goscribe` for programmatic access. The package wraps all internal functionality without exposing `internal/*` types. All long-running operations accept `context.Context`.

#### Creating a Client

```go
import (
    "context"
    "log"
    "time"

    "github.com/house/goscribe/pkg/goscribe"
)

// Option A: use a named provider from your goscribe config
client, err := goscribe.NewClient(
    goscribe.WithConfiguredProvider("ollama"),
    goscribe.WithModel("llama2"),
    goscribe.WithOutputDir("docs"),
)
if err != nil {
    log.Fatal(err)
}

// Option B: inject a custom Provider (good for testing or private backends)
type MyProvider struct{}
func (p MyProvider) Generate(ctx context.Context, prompt string) (string, error) {
    return "documentation from my backend", nil
}

client, err := goscribe.NewClient(
    goscribe.WithProvider(MyProvider{}),
    goscribe.WithOutputDir("docs"),
)
```

#### Functional Options Reference

| Option | What it does |
|--------|-------------|
| `WithProvider(p Provider)` | Inject a custom Provider implementation |
| `WithConfiguredProvider(name)` | Select a named provider from persisted config |
| `WithModel(model)` | Override the AI model |
| `WithOutputDir(dir)` | Set output directory (default: `"docs"`) |
| `WithConfigFile(path)` | Load config from a specific file |
| `WithTimeout(d)` | AI call timeout (default: `5 * time.Minute`) |
| `WithRetry(n, backoff)` | Retry count and initial backoff (default: 3 retries, 2s backoff) |
| `WithVerbose(bool)` | Enable verbose logging |

#### Generating Documentation

```go
result, err := client.Generate(ctx, "/path/to/project", goscribe.GenerateOptions{})
if err != nil {
    log.Fatal(err)
}
log.Printf("generated docs in %s (state saved: %v, commit: %s)",
    result.OutputDir, result.StateSaved, result.Commit)
```

`GenerateOptions` fields:

- `OutputDir string` - override client default for this run
- `Force bool` - regenerate even if output exists
- `SkipStateSave bool` - skip writing `.goscribe-state`

`GenerateResult` fields: `SourcePath`, `OutputDir`, `FilesGenerated`, `StateSaved`, `Commit`.

#### Updating Documentation

```go
result, err := client.Update(ctx, "/path/to/project", goscribe.UpdateOptions{})
if err != nil {
    log.Fatal(err)
}
if result.NoChanges {
    log.Println("no changes since last generation")
    return
}
log.Printf("updated %d files (%s..%s)", result.FilesUpdated, result.FromCommit, result.ToCommit)
```

`UpdateOptions` fields:

- `OutputDir string` - override client default
- `ChangedFiles []string` - supply explicit file list (empty = auto-detect from git)
- `FromCommit string` - override start commit (empty = load from `.goscribe-state`)
- `ToCommit string` - override end commit (empty = HEAD)
- `SkipStateSave bool` - skip writing updated state

`UpdateResult` fields: `SourcePath`, `OutputDir`, `FromCommit`, `ToCommit`, `ChangedFiles`, `FilesUpdated`, `NoChanges`, `StateSaved`.

#### Error Handling

All errors are classifiable with `errors.Is`:

```go
result, err := client.Update(ctx, ".", goscribe.UpdateOptions{})
if err != nil {
    switch {
    case errors.Is(err, goscribe.ErrStateNotFound):
        // need to run Generate first
    case errors.Is(err, goscribe.ErrNotGitRepository):
        // path is not a git repo
    case errors.Is(err, goscribe.ErrProviderNotConfigured):
        // no provider configured
    default:
        log.Fatal(err)
    }
}
```

Sentinel errors: `ErrInvalidContext`, `ErrInvalidPath`, `ErrProviderNotConfigured`, `ErrUnsupportedProvider`, `ErrGenerationFailed`, `ErrUpdateFailed`, `ErrNotGitRepository`, `ErrStateNotFound`, `ErrStateInvalid`.

The `goscribe.Error` struct wraps errors with `Op` (operation name), `Kind` (sentinel), `Path`, and `Err` (underlying cause). It supports `Unwrap()` and `Is()` for standard Go error inspection.

#### Library Version

```go
fmt.Println(goscribe.Version) // "1.0.0"
```

### 3. CI/CD Integration

CI mode produces machine-readable output instead of human-friendly text. Enable with the `--ci` flag.

#### CI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ci` | `false` | Enable CI mode (structured output, non-interactive) |
| `--output-format` | `text` | Output format: `text`, `json`, `markdown`, `github` |

#### Output Formats

Each format writes a `ci.Result` structure:

- **json** - full structured JSON with all fields
- **markdown** - formatted markdown with status, file counts, output dir
- **github** - GitHub Actions annotations (`::notice::`, `::error::`)
- **text** - plain text (same messages as non-CI mode)

JSON output example (successful generate):

```json
{
  "success": true,
  "command": "generate",
  "files_generated": 12,
  "output_dir": "docs"
}
```

JSON output example (update with no changes):

```json
{
  "success": true,
  "command": "update",
  "no_changes": true
}
```

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Generation or update failed |
| 2 | Update succeeded but no changes detected |

In CI mode, error results are still written to stdout before the non-zero exit. This lets CI systems capture the structured output regardless of the exit code.

#### Pipeline Usage

```yaml
- name: Generate docs
  env:
    GOSCRIBE_PROVIDER: openai
    GOSCRIBE_MODEL: gpt-4
  run: |
    goscribe generate ./src --ci --output-format json

- name: Update docs on PRs
  env:
    GOSCRIBE_PROVIDER: openai
    GOSCRIBE_MODEL: gpt-4
  run: |
    goscribe update ./src --ci --output-format github
```

Provider credentials can come from environment variables, a mounted config file, or the `provider add` command run in a prior step.

#### GitHub Action (Composite Action)

Location: `.github/actions/goscribe/action.yml`

The composite action installs GoScribe from a GitHub release, runs `generate`, and optionally commits the result.

**Inputs:**

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `path` | yes | `"."` | Source directory |
| `provider` | no | `"openai"` | AI provider (`openai` or `ollama`) |
| `model` | no | `"gpt-4"` | Model name |
| `output` | no | `"docs"` | Output directory |
| `api-key` | no | `""` | API key (use secrets for OpenAI) |
| `version` | no | `"latest"` | GoScribe version to install |
| `force` | no | `"false"` | Force regeneration |
| `verbose` | no | `"false"` | Verbose output |
| `commit-docs` | no | `"false"` | Commit generated docs |
| `commit-message` | no | `"docs: update documentation via goscribe"` | Commit message |

**Outputs:**

| Output | Description |
|--------|-------------|
| `output-dir` | Directory where docs were generated |

**Usage in a workflow:**

```yaml
steps:
  - uses: actions/checkout@v4

  - uses: ./.github/actions/goscribe
    with:
      path: "./src"
      provider: "openai"
      model: "gpt-4"
      api-key: ${{ secrets.OPENAI_API_KEY }}
      output: "docs"
      commit-docs: "true"
      commit-message: "docs: regenerate from CI"
```

The action downloads the binary from GitHub releases using the OS and architecture of the runner, so it works on `ubuntu-latest` and `macos-latest`. When `commit-docs` is `"true"`, it stages and commits the output directory if there are changes.

### 4. OpenCode Plugin

The `plugin/opencode/` package provides IDE/editor integration. It connects to the OpenCode agent system for real-time documentation generation as code changes.

#### Plugin Architecture

The plugin runs as a background service with its own lifecycle:

1. `NewClient(cfg)` creates a plugin client
2. `RegisterHandler()` attaches event callbacks
3. `Init(ctx)` starts the connection and optional file watcher
4. `Run(ctx)` blocks until shutdown
5. `Shutdown(ctx)` stops gracefully with a 5-second timeout

#### Configuration

Config file: `<project>/.opencode/goscribe.yaml`

```yaml
enabled: true
auto_generate: false
output_dir: "docs"
provider: "ollama"
model: "llama2"
project_path: "."
watch_patterns:
  - "**/*.go"
ignore_patterns:
  - "**/*_test.go"
  - "vendor/**"
agent_address: "http://localhost:8080"
```

`PluginConfig` fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Enabled` | bool | `true` | Enable the plugin |
| `AutoGenerate` | bool | `false` | Auto-generate on file changes |
| `OutputDir` | string | `"docs"` | Doc output directory |
| `ProviderName` | string | `""` | AI provider name |
| `Model` | string | `""` | Model name |
| `ProjectPath` | string | `""` | Project root |
| `WatchPatterns` | []string | `["**/*.go"]` | Glob patterns to watch |
| `IgnorePatterns` | []string | `["**/*_test.go", "vendor/**"]` | Patterns to ignore |
| `AgentAddress` | string | `"http://localhost:8080"` | Agent connection address |

Load with `LoadConfig(path)` or `LoadConfigFromProject(projectPath)`. Save with `cfg.Save(path)`.

#### Event System

| EventType | When | Use |
|-----------|------|-----|
| `EventInit` | Plugin initialized | Setup logging, notify IDE |
| `EventFileSaved` | A watched file changes | Trigger doc generation |
| `EventTrigger` | Explicit user trigger | Manual doc generation |
| `EventShutdown` | Plugin shutting down | Cleanup resources |

Register handlers:

```go
client := opencode.NewClient(cfg)
client.RegisterHandler(opencode.EventFileSaved, func(ctx context.Context, event opencode.Event) error {
    log.Printf("file changed: %s", event.FilePath)
    // trigger documentation generation
    return nil
})
```

#### File Watcher

`FileWatcher` monitors `.go` files with built-in filtering:

- Only watches `.go` files (excludes `_test.go`)
- Skips `vendor/`, `node_modules/`, `.git/`, and hidden directories
- Reads and respects `.gitignore` patterns
- Debounces rapid changes with a 2-second default (`SetDebounceDelay` to customize)

The watcher starts automatically during `Init()` when `AutoGenerate` is `true`. Failure to start the watcher does not prevent initialization (it logs the error and continues).

#### Plugin States

`StateUninitialized` -> `StateInitializing` -> `StateRunning` -> `StateShuttingDown` -> `StateStopped`

Query with `client.State()`. Plugin methods return errors if called in the wrong state.

#### Generation Types

- `GenerationRequest`: `SourcePath`, `OutputDir`, `Force`
- `GenerationResponse`: `Success`, `OutputDir`, `FilesGenerated`, `Error`

### Docker

Multi-stage build targeting minimal production images.

#### Build

```bash
docker build -t goscribe .
```

#### Run

```bash
# Print version
docker run --rm goscribe version

# Generate docs for a mounted project
docker run --rm -v $(pwd):/src goscribe generate /src

# With provider config
docker run --rm -v ~/.goscribe.yaml:/home/goscribe/.goscribe.yaml -v $(pwd):/src goscribe generate /src

# CI/CD with env vars
docker run --rm \
  -e GOSCRIBE_PROVIDER=openai \
  -e GOSCRIBE_MODEL=gpt-4 \
  -v $(pwd):/src \
  goscribe generate /src --ci --output-format json
```

The image runs as a non-root user (`goscribe`, uid 1000) on Alpine 3.21 with `ca-certificates` and `git` installed. Entrypoint is `goscribe` with default command `--help`.

## Development Notes

- Use `go mod tidy` after adding imports.
- All business logic lives in `internal/` packages.
- Keep `cmd/` files focused on CLI concerns only.
- Provider interface allows easy addition of new AI backends. Add a case to the switch in `provider.go` and create a new file in `internal/ai/`.
- The `.goscribe-state` file should be added to `.gitignore` in target projects (it's project-specific state, not source).
- GoReleaser ldflags inject `version.Version` at release. Don't rely on the hardcoded value for release builds.
- When adding new usage modes (library, plugin), maintain the provider interface as the common abstraction.

## Scooby-Doo Quotes

Commit messages from `bump` command include random Scooby-Doo quotes. Quote selection is deterministic based on version string (hash-based, not random). Eight quotes available.
