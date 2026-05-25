# GoScribe OpenCode Plugin

This example shows how to use the GoScribe OpenCode plugin to generate documentation inside an IDE or editor environment.

The plugin watches your Go source files and dispatches events when files change. You can register handlers to trigger documentation generation in response to these events.

## Prerequisites

- Go 1.24 or later
- GoScribe installed and configured with an AI provider

## Setup

### 1. Install GoScribe

```bash
git clone https://github.com/house/goscribe.git
cd goscribe
go build ./cmd/goscribe
```

### 2. Configure an AI provider

```bash
# For local Ollama
goscribe provider add ollama --url http://localhost:11434 --model llama2

# For OpenAI
goscribe provider add openai --key sk-... --model gpt-4
```

### 3. Create plugin config

Copy the example config into your project:

```bash
mkdir -p .opencode
cp examples/plugin/.opencode/goscribe.yaml .opencode/goscribe.yaml
```

Edit the config to match your setup. At minimum, set `provider` and `model` to match what you configured in step 2.

### 4. VS Code integration

The plugin works with any editor that can run Go programs. For VS Code:

1. Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go)
2. Open your project folder
3. Run the plugin example with `go run main.go` in a VS Code terminal
4. When `auto_generate: true` is set, saving any `.go` file dispatches a `file_saved` event. Register a handler to trigger documentation generation in response.

You can also wire it as a VS Code task. Add this to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "goscribe: watch",
      "type": "shell",
      "command": "go run main.go",
      "isBackground": true,
      "problemMatcher": []
    }
  ]
}
```

Then run it from the command palette: `Tasks: Run Task` and pick `goscribe: watch`.

## Configuration Options

All options go in `.opencode/goscribe.yaml`:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Turn the plugin on or off |
| `auto_generate` | bool | `false` | Generate docs on every file save |
| `output_dir` | string | `docs` | Where to write generated docs |
| `provider` | string | (from `~/.goscribe.yaml`) | AI provider name |
| `model` | string | (from `~/.goscribe.yaml`) | AI model name |
| `project_path` | string | `.` | Root directory of the Go project |
| `watch_patterns` | list | `["**/*.go"]` | Globs for files to watch |
| `ignore_patterns` | list | `["**/*_test.go", "vendor/**"]` | Globs for files to skip |
| `agent_address` | string | `http://localhost:8080` | OpenCode agent address |

### Minimal config

If you just want the defaults with a specific provider:

```yaml
enabled: true
provider: ollama
model: llama2
```

### Full config with auto-generation

```yaml
enabled: true
auto_generate: true
output_dir: docs
provider: openai
model: gpt-4
project_path: .
watch_patterns:
  - "**/*.go"
ignore_patterns:
  - "**/*_test.go"
  - "vendor/**"
  - "internal/mock/**"
agent_address: "http://localhost:8080"
```

## Event Types

The plugin dispatches these events:

| Event | When |
|-------|------|
| `init` | Plugin finishes initialization |
| `file_saved` | A watched file changes (debounced, 2s delay) |
| `trigger` | Explicit generation request |
| `shutdown` | Plugin begins graceful shutdown |

Register handlers with `client.RegisterHandler(eventType, handler)`. See `main.go` in this directory for a working example.

## Running the Example

From this directory:

```bash
go run main.go
```

The program loads `.opencode/goscribe.yaml`, registers handlers for each event type, and blocks until you press Ctrl+C.

If `auto_generate` is enabled in the config, edit and save any `.go` file in the project. The `file_saved` handler prints the changed file path.

## Troubleshooting

### "plugin is disabled in configuration"

The `enabled` field in your config is set to `false`. Open `.opencode/goscribe.yaml` and set `enabled: true`.

### Config file not found

The plugin looks for `.opencode/goscribe.yaml` relative to the working directory. Make sure you run the program from the project root, or set `project_path` to the correct absolute path.

### "plugin already initialized"

`Init` was called twice. The client can only be initialized once. If you need to restart, call `Shutdown` first, then create a new `Client`.

### No events on file save

Check these things:

- `auto_generate` must be `true` for file watching to start
- The file must match a pattern in `watch_patterns` and not match anything in `ignore_patterns`
- Only `.go` files are watched by default
- Test files (`*_test.go`) and files in `vendor/` are always excluded

### Provider not found

The `provider` value in your plugin config must match a provider name set up with `goscribe provider add`. Run `goscribe provider list` to see what's configured.

### Generated docs are empty

The AI provider might not be reachable. Test it directly:

```bash
goscribe generate -v
```

Verbose output (`-v`) shows the provider response. If that works, the plugin should work too. Check that `provider` and `model` in `.opencode/goscribe.yaml` match your `~/.goscribe.yaml` settings.
