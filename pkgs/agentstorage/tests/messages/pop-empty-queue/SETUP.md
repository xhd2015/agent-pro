# Scenario

**Feature**: pop on empty message queue

```
no AppendMessage -> PopMessages -> empty slice, no error
```

## Preconditions

- Session has no queued messages.
- `messages.jsonl` may not exist yet.

## Steps

1. Set `req.Action = "pop_empty"`.
2. Call `PopMessages` without prior appends.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "pop_empty"
	return nil
}
```