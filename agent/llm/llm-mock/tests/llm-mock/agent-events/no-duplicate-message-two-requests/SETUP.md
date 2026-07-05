# Scenario

**Feature**: genStream think+message breakpoint must not duplicate message within one serve

With breakpoint dequeue, each HTTP request consumes leading think(s) plus one message breakpoint.
Agent-events must log exactly that consumed slice per serve (no peek-ahead duplicate message).

```
POST #1 user:Hello -> think+message breakpoint -> agent-events: think, message
POST #2 user:Hello -> renewed genStream breakpoint -> agent-events: think, message
```

## Steps

1. Empty `exchanges[]` (random fallback).
2. Two identical single-message requests.
3. Read `--agent-events-file` after both requests.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"Hello"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"Hello"}]}`,
	}
	return nil
}
```