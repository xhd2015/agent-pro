# Scenario

**Feature**: reverse conversion from canonical AgentEvents to grok session updates

```
# canonical events become typed session updates, then JSONL wire lines
[]types.AgentEvent -> ToSession -> []SessionUpdate -> ToWireLines -> []string
```

## Preconditions

- `ToSession` maps ActionMessage/ActionThink/ActionToolCall/ActionDone to ACP updates.
- `ToWireLines` marshals updates; `req.ToOpts` controls flat vs envelope output.
- Each leaf sets `req.Target = "to_session"` and populates `req.Events`.

## Steps

1. Set `req.Target = "to_session"`.
2. Leaf SETUPs build canonical `req.Events` with grok_session extensions where needed.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "to_session"
	return nil
}
```