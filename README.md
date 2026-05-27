# agent-traces

Standalone web UI for headless agent execution traces.

## Usage

```sh
go run ./ --port 9898 --open
```

By default the viewer reads traces from `~/.knowledge-hub/agent-traces`.
Use `--data-dir` to point at a different knowledge-hub data directory:

```sh
go run ./ --data-dir /path/to/.knowledge-hub
```

To mount the app under a sub-route, pass `--route-prefix`:

```sh
go run ./ --route-prefix agent-traces
```

Then open `http://localhost:<port>/agent-traces/`.

## Development

### Dev Mode

Dev mode starts Vite automatically for frontend hot-reload and proxies requests through the Go backend:

```sh
go run ./ --dev
```

Dev mode also supports route prefixes:

```sh
go run ./ --dev --route-prefix my-app
```

The Go server starts and proxies to a Vite dev server on a free port, so you get instant hot-reload for React changes.

### Release Mode (Production Build)

1. Build the frontend:
```sh
go run ./script/build
```

2. Run the server (embeds the built frontend):
```sh
go run ./
```

### Build Binary

```sh
go run ./script/build && go build -o /tmp/agent-traces ./
```
