# Scenario

**Feature**: UnwrapEvent drops a 3-level SSE line whose outer type is unknown

## Preconditions
- Input has an outer type that is NOT one of `"message"`, `"agent_event"`, or `"run_complete"`.
- Inner payload structure is irrelevant.

## Steps
1. Set `SSEInput` to a 3-level event with `"lsp_event"` type.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SSEInput = `{"type":"lsp_event","payload":{"type":"updated","payload":{"key":"value"}}}`
	return nil
}
```
