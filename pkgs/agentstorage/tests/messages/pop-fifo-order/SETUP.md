# Scenario

**Feature**: pop messages drains queue in FIFO order

```
AppendMessage("oldest") + AppendMessage("middle") + AppendMessage("newest") -> PopMessages -> FIFO
```

## Preconditions

- Three messages appended in known order.
- Single `PopMessages` drains the entire queue.

## Steps

1. Set `req.Action = "pop_fifo"`.
2. Append three messages in chronological order.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "pop_fifo"
	req.MessageText = []string{"oldest", "middle", "newest"}
	return nil
}
```