# Scenario

**Feature**: read events on session with no log file

```
ReadEvents(new session, offset=0) -> empty slice, offset 0, no error
```

## Preconditions

- Session id has never received an `AppendEvent`.
- No `events.jsonl` exists yet.

## Steps

1. Set `req.Action = "read_empty"`.
2. Read events without prior append.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "read_empty"
	req.SessionID = "sess_empty_events"
	return nil
}
```