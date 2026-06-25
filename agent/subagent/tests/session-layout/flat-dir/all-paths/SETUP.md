# Scenario

**Feature**: default flat layout writes all standard subagent artifacts

```
# events, messages, questions/, progress/ under SessionLayout.Dir
subagent.Run -> events.jsonl + messages.jsonl + questions/ + progress/
```

## Preconditions

- Default layout flags (messages, questions, progress all enabled).

## Steps

1. Inherit flat-dir base setup with all features enabled.

## Context

- `pid` is removed after run completes.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.SessionDir == "" {
		configureFlatDirBase(t, req)
	}
	return nil
}
```
