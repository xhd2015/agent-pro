# Scenario

**Feature**: auth status --app validates app-level token via apps.connections.open

```
Caller -> slack-msg auth status --app [options] -> apps.connections.open -> app status
```

## Preconditions

- App token required (`--app-token` / `SLACK_APP_TOKEN` / config `appToken`).
- Unit leaves use slacktest default `apps.connections.open` (ok + websocket url).

## Steps

1. Leaves include `--app` in args.
2. Success leaves attach default slacktest via `SlackAPIURL`.

## Context

- Contract: validate with **`apps.connections.open`** (Socket Mode / connections oriented).
- Human includes fixed `note:` line about app-level token usage.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
