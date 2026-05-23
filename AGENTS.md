# AGENTS.md

> GoScribe - AI-powered documentation generator for Go projects.

## Project Overview

- **Name:** goscribe
- **Language:** Go
- **Type:** CLI application
- **Purpose:** Reads application source code and produces comprehensive user documentation using AI providers (OpenAI, Ollama)
- **Module:** `github.com/house/goscribe`

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
│   │   ├── provider.go      # Provider interface + factory
│   │   ├── openai.go        # OpenAI client implementation
│   │   └── ollama.go        # Ollama client implementation
│   ├── config/
│   │   └── config.go        # Configuration loading/saving
│   ├── docs/
│   │   ├── generator.go     # Documentation generation engine
│   │   ├── updater.go       # Incremental doc updates
│   │   └── state.go         # Git commit state tracking
│   ├── git/
│   │   └── git.go           # Git operations wrapper
│   └── version/
│       └── version.go       # Version management + Scooby-Doo quotes
└── pkg/
    └── providers/
        └── providers.go     # Provider configuration persistence
```

### Key Design Decisions

- **cmd/** contains Cobra commands only - no business logic
- **internal/** contains all implementation details (compiler-enforced privacy)
- **pkg/providers/** is the only public package (if needed for reuse)
- Configuration stored in `~/.goscribe.yaml`
- Git state tracked in `.goscribe-state` file in source directory

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
# Generate docs for current directory
./goscribe generate

# Generate docs for specific path
./goscribe generate /path/to/project

# Force regeneration
./goscribe generate -f
```

### Update Documentation (Git-aware)
```bash
# Update docs based on changes since last generation
./goscribe update

# Requires git repository and previous .goscribe-state file
```

### Provider Management
```bash
# Add Ollama provider
./goscribe provider add ollama --url http://localhost:11434 --model llama2

# Add OpenAI provider
./goscribe provider add openai --key sk-... --model gpt-4

# List providers
./goscribe provider list

# Remove provider
./goscribe provider remove ollama
```

### Version Management
```bash
# Show version
./goscribe version

# Bump patch version (0.0.x), commit with Scooby-Doo quote
./goscribe bump

# Tag current commit with version
./goscribe tag
```

## Configuration

Config file: `~/.goscribe.yaml`

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

Environment variables (prefix: `GOSCRIBE_`):
- `GOSCRIBE_PROVIDER`
- `GOSCRIBE_MODEL`
- `GOSCRIBE_OUTPUT`
- `GOSCRIBE_VERBOSE`

## Git Integration

- **First run:** `generate` saves current commit hash to `.goscribe-state`
- **Updates:** `update` compares current commit with saved state, only processes changed files
- **Requirements:** Source directory must be a git repository

## Scooby-Doo Quotes

Commit messages from `bump` command include random Scooby-Doo quotes:
- "Ruh-roh! Looks like we got a new version, Raggy!"
- "Scooby-Dooby-Doo! Version bump time!"
- "Zoinks! That's one spooky version update!"
- etc.

## Development Notes

- Use `go mod tidy` after adding imports
- All business logic in `internal/` packages
- Keep `cmd/` files focused on CLI concerns only
- Provider interface allows easy addition of new AI backends

## Testing

- No tests exist yet — add `*_test.go` files alongside implementation
- Run tests: `go test ./...`
- Run vet: `go vet ./...`

## Known Limitations

- Documentation generation requires active AI provider
- Git integration requires initialized git repository
- State tracking is file-based (not database)
- No tests written yet
