# Scenario

**Feature**: tail output semantically matches grok_session.FromUpdatesJSONL

```
# same wire fixture through both paths
wire lines -> TailUpdatesFromOffset events ≡ grok_session.FromUpdatesJSONL(wire lines)
```

## Preconditions

- `grok_session` package is implemented and GREEN (reference semantics).
- Tail path must delegate to grok_session for parity (RED before refactor).

## Steps

1. Set `req.Target = "integration"`.
2. Leaf SETUPs provide a representative full-turn wire fixture.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Target = "integration"
	req.UpdatesPath = newTempUpdatesPath(t)
	return nil
}
```