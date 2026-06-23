# Scenario

**Feature**: full human-readable log for a Codex session

```
# resolve rollout path by session UUID
sessions.Find(codexHome, sessionID) -> rollout JSONL path

# normalize events and print compact trace lines
sessions.PrintLog(path, w, tail) -> RUN, ASSISTANT, EDIT, REASONING blocks
```

## Preconditions

- This branch tests the `log` operation (`--log` mode).
- Codex trace adapter is registered in root Setup.

## Steps

1. Set `req.Operation = "log"`.
2. Leaf Setup writes rollout JSONL with displayable and skipped events.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "log"
	return nil
}
```