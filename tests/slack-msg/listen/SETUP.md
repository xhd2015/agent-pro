# Scenario

**Feature**: slack-msg listen Socket Mode inbound bridge

```
Caller -> slack-msg listen [options] -> Socket Mode -> filter -> agent-run -> PostMessage reply
```

## Preconditions

- Subcommand always `listen` (via harness `ListenMode` + `defaultListenArgs`).
- Bot flag is `--token` (not legacy `--bot-token`).

## Steps

1. Mark `req.ListenMode = true` so `Run` builds listen argv with `--token` / `--app-token`.
2. Isolate workdir.

## Context

- Port of `tests/slack-listen` with binary name `slack-msg` and unified bot flag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ListenMode = true
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
