# Scenario

**Feature**: `--resume-from-grok-session` for a Grok id not present under GROK_HOME fails

```
# empty GROK_HOME sessions tree; id never seeded
agent-run run --resume-from-grok-session UUID
  -> exit 1; not found
```

## Preconditions

- Isolated `GROK_HOME` exists but has no `summary.json` for the requested id.
- No agent-run meta mapping for the id (store empty).

## Steps

1. Use default `req.GrokSessionID` without calling `seedGrokSession`.
2. Run with flag only (no `--agent-runner` → treat as grok-tty once implemented).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// GROK_HOME is empty of sessions; id is still a well-formed UUID.
	req.Args = runArgs(req, req.GrokSessionID)
	return nil
}
```
