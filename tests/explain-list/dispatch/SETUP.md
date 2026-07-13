# Scenario

**Feature**: list dispatch never calls the agent runner

```
# EXPLAIN_AGENT_PATH = failing fake; explain list -> exit 0, fake never runs
```

## Preconditions

- Fake agent stub exits 99 with `FAKE_AGENT_INVOKED` if executed.
- Root Setup always sets `EXPLAIN_AGENT_PATH` to that stub.

## Steps

1. Seed at least one session so list has work to do (not only empty-path).
2. Run list; assert fake not invoked.

## Context

- Guards against accidental fall-through into ask/resume LLM path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("dispatch setup: explain binary not built")
	}
	if req.FakeAgentPath == "" {
		t.Fatalf("dispatch setup: fake agent path missing")
	}
	return nil
}
```
