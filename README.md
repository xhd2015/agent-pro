# agent-traces

Standalone web UI for headless agent execution traces.

## Usage

```sh
go run ./cmd/agent-traces --port 9898 --open
```

By default the viewer discovers trace roots from:

- `~/.agent-traces`
- `~/.*/agent-traces`
- `./.agent-traces`
- `./.*/agent-traces`

Use `--data-dir` to point at a specific knowledge-hub style data directory:

```sh
go run ./cmd/agent-traces --data-dir /path/to/.knowledge-hub
```

You can also pass one source argument:

```sh
# Single JSONL event file
go run ./cmd/agent-traces /path/to/events.jsonl

# One trace session directory
go run ./cmd/agent-traces /path/to/agent-traces/20260602-123456.000000

# A root directory containing multiple trace session directories
go run ./cmd/agent-traces /path/to/agent-traces
```

To mount the app under a sub-route, pass `--route-prefix`:

```sh
go run ./cmd/agent-traces --route-prefix agent-traces
```

Then open `http://localhost:<port>/agent-traces/`.

## Development

### Dev Mode

Dev mode starts Vite automatically for frontend hot-reload and proxies requests through the Go backend:

```sh
go run ./cmd/agent-traces --dev
```

Dev mode also supports route prefixes:

```sh
go run ./cmd/agent-traces --dev --route-prefix my-app
```

The Go server starts and proxies to a Vite dev server on a free port, so you get instant hot-reload for React changes.

### Release Mode (Production Build)

1. Build the frontend:
```sh
go run ./script/build
```

2. Run the server (embeds the built frontend):
```sh
go run ./cmd/agent-traces
```

### Build Binary

```sh
go run ./script/build && go build -o /tmp/agent-traces ./cmd/agent-traces
```
