# Scenario

**Feature**: `run` without `--json` uses human-readable output

```
agent-run run --agent-runner fake-codex "hi" → stdout not all JSON object lines
```

## Preconditions

- `fake-codex` on PATH (inherited from root `SETUP.md`).
- Grouping `Setup` prefixes `req.Args` with `run --agent-runner fake-codex`.

## Steps

1. Run `agent-run run --agent-runner fake-codex "hi"` without `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "hi")
	return nil
}
```