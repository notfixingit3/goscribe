# GoScribe

Generate comprehensive documentation from Go source code using AI.

GoScribe reads your project's source, sends it to an AI provider, and produces full user documentation with real examples. It supports 17 AI providers including OpenAI, Anthropic Claude, Google Gemini, xAI Grok, NVIDIA NIM, GitHub Copilot, and local models via Ollama, vLLM, and LM Studio. It tracks git state so you can incrementally update docs when your code changes, and supports parallel generation with content-addressed caching for performance.

## Install

Build from source:

```bash
git clone https://github.com/house/goscribe.git
cd goscribe
go build ./cmd/goscribe
```

Move the binary somewhere on your PATH:

```bash
mv goscribe /usr/local/bin/
```

## Quick Start

```bash
# Add an AI provider
goscribe provider add ollama --url http://localhost:11434 --model llama2

# Generate docs for the current directory
goscribe generate

# Later, after code changes
goscribe update
```

Documentation lands in `./docs` by default.

## Commands

### `goscribe generate [path]`

Read source code and produce documentation.

```bash
goscribe generate                    # current directory
goscribe generate /path/to/project   # specific project
goscribe generate -f                 # force, overwrite existing docs
```

If the path is a git repo, GoScribe saves the current commit hash to `.goscribe-state`. This lets `update` find what changed later.

Flags:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | `-f` | false | Regenerate even if docs exist |
| `--provider` | `-p` | config | AI provider to use |
| `--model` | `-m` | config | Model to use |
| `--output` | `-o` | `docs` | Output directory |
| `--verbose` | `-v` | false | Verbose output |
| `--ci` | | false | CI mode: structured output, exit codes |
| `--output-format` | | `text` | Output format: text, json, markdown, github |
| `--timeout` | | `5m` | Timeout for AI calls |
| `--retries` | | `3` | Retries for transient AI errors |
| `--retry-backoff` | | `2s` | Initial backoff between retries |
| `--workers` | | `1` | Concurrent workers for generation |
| `--cache-dir` | | `""` | Content-addressed cache directory |
| `--profile` | | `""` | Documentation profile (see Profiles below) |
| `--template` | | `""` | Documentation template (see Templates below) |

### `goscribe update [path]`

Update docs based on git changes since the last `generate`.

```bash
goscribe update
goscribe update /path/to/project
```

This requires a git repository and a `.goscribe-state` file from a previous `generate` run. If nothing changed, it tells you so.

### `goscribe provider`

Manage AI providers.

```bash
# Add Ollama
goscribe provider add ollama --url http://localhost:11434 --model llama2

# Add OpenAI
goscribe provider add openai --key sk-... --model gpt-4

# Set a default
goscribe provider add ollama --url http://localhost:11434 --model llama2 --default

# List configured providers
goscribe provider list

# Remove one
goscribe provider remove ollama
```

Supported providers: `openai`, `ollama`, `claude`, `gemini`, `kimi`, `openrouter`, `lmstudio`, `llamacpp`, `ollama-cloud`, `ollama-selfhosted`, `opencode`, `xai`, `nvidia`, `github-copilot`, `zai`, `vllm`

Subcommands:

| Command | Description |
|---------|-------------|
| `provider add [name]` | Add a provider (flags: `--url`, `--key`, `--model`, `--default`) |
| `provider list` | Show all configured providers |
| `provider remove [name]` | Remove a provider |

### `goscribe version`

Print the current version.

```bash
goscribe version
```

### `goscribe bump`

Bump the patch version (0.0.x), stage all files, and commit with a random Scooby-Doo quote.

```bash
goscribe bump
```

### `goscribe tag`

Tag the current commit with the version number.

```bash
goscribe tag
```

## Library Usage

GoScribe exposes a public API in `pkg/goscribe` for programmatic use within the same module.

```go
import (
    "context"
    "log"

    "github.com/house/goscribe/pkg/goscribe"
)

func main() {
    client, err := goscribe.NewClient(
        goscribe.WithConfiguredProvider("ollama"),
        goscribe.WithModel("llama2"),
        goscribe.WithOutputDir("docs"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Generate documentation for a project
    result, err := client.Generate(ctx, "/path/to/project", goscribe.GenerateOptions{})
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Generated %d files in %s", result.FilesGenerated, result.OutputDir)

    // Later, update only changed files
    updateResult, err := client.Update(ctx, "/path/to/project", goscribe.UpdateOptions{})
    if err != nil {
        log.Fatal(err)
    }
    if updateResult.NoChanges {
        log.Println("No changes since last generation")
    } else {
        log.Printf("Updated %d files", updateResult.FilesUpdated)
    }
}
```

The `pkg/goscribe` package provides `NewClient` with functional options (`WithProvider`, `WithModel`, `WithOutputDir`, `WithTimeout`, `WithRetry`, `WithWorkers`, `WithCacheDir`, etc.), context-aware `Generate` and `Update` methods, and structured errors compatible with `errors.Is`.

### Parallel Generation

Use `WithWorkers(n)` to process files concurrently. This is especially useful for large codebases:

```go
client, err := goscribe.NewClient(
    goscribe.WithConfiguredProvider("openai"),
    goscribe.WithWorkers(5),
    goscribe.WithCacheDir("/tmp/goscribe-cache"),
)
```

### Content-Addressed Caching

Enable caching with `WithCacheDir` to skip AI calls for unchanged files. The cache uses SHA-256 content hashes, so identical source always produces identical documentation without calling the provider:

```go
client, err := goscribe.NewClient(
    goscribe.WithConfiguredProvider("openai"),
    goscribe.WithCacheDir("/tmp/goscribe-cache"),
)
```

For provider configuration persistence, use `pkg/providers`:

```go
import "github.com/house/goscribe/pkg/providers"

providers.SaveProvider(providers.ProviderConfig{
    Name:    "ollama",
    URL:     "http://localhost:11434",
    Model:   "llama2",
    Default: true,
})
```

## OpenCode Plugin

GoScribe ships with an OpenCode plugin for IDE integration. It watches your source files and dispatches events on file changes, which you can handle to trigger documentation generation.

### Setup

Create a config file at `.opencode/goscribe.yaml` in your project:

```yaml
enabled: true
auto_generate: false
output_dir: docs
provider: ollama
model: llama2
project_path: .
watch_patterns:
  - "**/*.go"
ignore_patterns:
  - "**/*_test.go"
  - "vendor/**"
agent_address: http://localhost:8080
profile: software-documenter
```

### Using the Plugin

```go
package main

import (
    "context"
    "log"

    "github.com/house/goscribe/plugin/opencode"
)

func main() {
    cfg, err := opencode.LoadConfigFromProject(".")
    if err != nil {
        log.Fatal(err)
    }

    client := opencode.NewClient(cfg)

    // Register event handlers
    client.RegisterHandler(opencode.EventFileSaved,
        func(ctx context.Context, e opencode.Event) error {
            log.Printf("File saved: %s", e.FilePath)
            return nil
        },
    )

    if err := client.Init(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer client.Shutdown(context.Background())

    // Blocks until context is cancelled
    client.Run(context.Background())
}
```

The plugin dispatches events for file saves, manual triggers, initialization, and shutdown. Set `auto_generate: true` to start the file watcher, which will dispatch `EventFileSaved` events when Go files change. Register a handler to trigger documentation generation in response to events.

## CI/CD Integration

Pass `--ci` to get structured output and proper exit codes for pipelines.

```bash
# Basic CI mode
goscribe generate --ci

# With JSON output for parsing
goscribe generate --ci --output-format json

# GitHub-flavored markdown annotations
goscribe generate --ci --output-format github
```

Exit codes in CI mode:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | No changes detected (update only) |

Environment variables work well in CI. Set `GOSCRIBE_PROVIDER`, `GOSCRIBE_MODEL`, and `GOSCRIBE_OUTPUT` instead of a config file.

## GitHub Action

GoScribe provides a composite action at `.github/actions/goscribe/`. Add it to any workflow:

```yaml
jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Generate documentation
        uses: ./.github/actions/goscribe
        with:
          path: "."
          provider: "openai"
          model: "gpt-4"
          api-key: ${{ secrets.OPENAI_API_KEY }}
          output: "docs"
          commit-docs: "true"
          commit-message: "docs: update documentation via goscribe"
```

Action inputs:

| Input | Default | Description |
|-------|---------|-------------|
| `path` | `.` | Source directory |
| `provider` | `openai` | AI provider (see supported list below) |
| `model` | `gpt-4` | Model name |
| `api-key` | | Provider API key (use secrets) |
| `output` | `docs` | Output directory |
| `version` | `latest` | GoScribe version to install |
| `force` | `false` | Force regeneration |
| `verbose` | `false` | Verbose output |
| `commit-docs` | `false` | Commit generated docs automatically |
| `commit-message` | `docs: update documentation via goscribe` | Commit message |

The action downloads a release binary, runs `goscribe generate`, and optionally commits the results back to the repo.

## Docker

Build and run with Docker:

```bash
# Build the image
docker build -t goscribe .

# Run a command
docker run --rm goscribe version

# Generate docs for a local project
docker run --rm -v $(pwd):/src goscribe generate /src

# Mount your config file
docker run --rm \
  -v ~/.goscribe.yaml:/home/goscribe/.goscribe.yaml \
  -v $(pwd):/src \
  goscribe generate /src

# Use environment variables for CI
docker run --rm \
  -e GOSCRIBE_PROVIDER=openai \
  -e GOSCRIBE_MODEL=gpt-4 \
  -v $(pwd):/src \
  goscribe generate /src
```

The image uses a multi-stage build. The runtime image is based on Alpine and runs as a non-root user.

## Configuration

GoScribe looks for a config file at `~/.goscribe.yaml`. You can also pass `--config /path/to/file` on any command.

```yaml
provider: ollama
model: llama2
output: docs
verbose: false
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

| Variable | Maps to |
|----------|---------|
| `GOSCRIBE_PROVIDER` | `provider` |
| `GOSCRIBE_MODEL` | `model` |
| `GOSCRIBE_OUTPUT` | `output` |
| `GOSCRIBE_VERBOSE` | `verbose` |
| `GOSCRIBE_TIMEOUT` | `timeout` |
| `GOSCRIBE_RETRIES` | `retries` |
| `GOSCRIBE_RETRY_BACKOFF` | `retry_backoff` |
| `GOSCRIBE_PROFILE` | `profile` |
| `GOSCRIBE_TEMPLATE` | `template` |

Priority order: command flags > environment variables > config file > defaults.

## Documentation Templates

Templates control the **style** and **presentation** of generated documentation. They answer "how should this look" while profiles answer "what should this document." Templates are optional — without one, GoScribe uses a neutral writing style.

Use the `--template` flag or set `template` in your config file.

| Template | Description |
|----------|-------------|
| `elegant` | Sophisticated, polished, professional tone |
| `technical` | Precise, dense, code-heavy documentation |
| `futuristic` | Modern, forward-looking, innovative tone |
| `minimal` | Bare bones, just the facts, no embellishment |
| `friendly` | Conversational, welcoming, beginner-friendly |

### Using Templates

CLI:
```bash
goscribe generate --template elegant
goscribe update --template minimal
```

Config file:
```yaml
template: elegant
```

Library API:
```go
client, err := goscribe.NewClient(
    goscribe.WithConfiguredProvider("openai"),
    goscribe.WithTemplate("elegant"),
    goscribe.WithOutputDir("docs"),
)
```

Plugin config:
```yaml
template: friendly
```

Templates compose with profiles for even more control:
```bash
goscribe generate --profile github-readme-expert --template elegant
```

## Documentation Profiles

Profiles tailor the documentation style for different audiences and use cases. Use the `--profile` flag or set `profile` in your config file.

| Profile | Description |
|---------|-------------|
| `software-documenter` | General software documentation with examples and explanations (default) |
| `technical-writer` | Clear, concise technical writing with usage examples |
| `github-readme-expert` | GitHub README style with badges, installation, and quick start |
| `github-wiki-expert` | GitHub Wiki style with structured pages and cross-references |
| `api-reference` | API reference style with endpoints, parameters, and response formats |
| `developer-onboarding` | Onboarding guide for new developers joining the project |
| `operations-runbook` | Operations runbook with deployment, monitoring, and troubleshooting |

### Using Profiles

CLI:
```bash
goscribe generate --profile technical-writer
goscribe update --profile api-reference
```

Config file:
```yaml
profile: github-readme-expert
```

Library API:
```go
client, err := goscribe.NewClient(
    goscribe.WithConfiguredProvider("openai"),
    goscribe.WithProfile("technical-writer"),
    goscribe.WithOutputDir("docs"),
)
```

Plugin config:
```yaml
profile: developer-onboarding
```

## Git Integration

GoScribe uses git to track what changed between documentation runs.

1. `generate` saves the current commit hash to `.goscribe-state` in the source directory.
2. `update` reads that hash, compares it to HEAD, finds changed files, and rewrites only the affected documentation sections.
3. After updating, it saves the new commit hash.

This means you can run `goscribe generate` once, then `goscribe update` after each batch of changes. No need to regenerate everything from scratch.

## Requirements

- Go 1.24 or later
- An active AI provider (see supported providers above)
- A git repository (for `update`; `generate` works without git)

## Development

```bash
# Build
go build ./cmd/goscribe

# Run tests
go test ./...

# Vet
go vet ./...
```

Project layout:

```
cmd/
  goscribe/main.go   # Entry point
  root.go            # Root command, config setup
  generate.go        # generate command
  update.go          # update command
  provider.go        # provider subcommands
  version.go         # version, bump, tag commands
internal/
  ai/                # Provider interface, 12+ provider implementations
  config/            # Config loading and saving
  docs/              # Generation engine, updater, state tracking, cache
  git/               # Git operations
  version/           # Version management
pkg/
  goscribe/          # Public library API
  providers/         # Provider config persistence (public)
plugin/
  opencode/          # OpenCode IDE plugin client
```

## License

MIT
