# Scenario

**Feature**: `sessions --json` on empty store returns valid empty JSON

```
agent-run sessions --json (empty AGENT_RUN_HOME) → valid JSON, no sessions
```

## Preconditions

- Fresh `AGENT_RUN_HOME` has no sessions.

## Steps

1. Run `agent-run sessions --json` on an empty `AGENT_RUN_HOME`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "--json")
	return nil
}
```