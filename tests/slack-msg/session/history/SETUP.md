# Scenario

**Feature**: session history prints local messages.jsonl

```
slack-msg session history [options]
  -> resolve session id -> read messages.jsonl -> oldest→newest lines
```

## Preconditions

- Action is `history` as second arg.
- Prefers local log (no live Slack required for success leaves).

## Steps

1. Leaves set args starting with `"session", "history"`.
2. Success leaves seed map + messages.jsonl under HomeDir.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
