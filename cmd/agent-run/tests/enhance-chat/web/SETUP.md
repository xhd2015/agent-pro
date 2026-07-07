# Scenario

**Feature**: web SSE delivers full event timeline including think and error types

```
POST grok-tty -> GET events/stream tails events.jsonl
SSE must not filter to message-only rows
```

## Preconditions

- SSE uses `pkgs/agentevents.WatchEvents` file tail with byte-offset resume.

## Steps

1. Grouping setup sets `req.Area = "web"` and `req.Mode = "sse"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Area = "web"
	req.Mode = "sse"
	return nil
}
```