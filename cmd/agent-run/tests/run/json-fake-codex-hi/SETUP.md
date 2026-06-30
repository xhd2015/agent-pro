# Scenario

**Feature**: `run --json` streams NDJSON events from fake-codex

```
agent-run run --json --agent-runner fake-codex "hi" → stdout NDJSON, last event type done
```

## Preconditions

- `fake-codex` on PATH (inherited from root `SETUP.md`).
- Grouping `Setup` prefixes `req.Args` with `run --agent-runner fake-codex`.

## Steps

1. Run `agent-run run --json --agent-runner fake-codex "hi"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "--json", "hi")
	return nil
}
```