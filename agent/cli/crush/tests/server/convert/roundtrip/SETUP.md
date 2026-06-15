## Preconditions
- These tests exercise `crush_types.ToCrush` and `crush_types.FromCrush` directly.
- No server or external dependency needed.
- Each leaf provides a JSON array of `types.AgentEvent` as `req.AgentEventsJSON`.

## Steps
1. Set `req.Mode = "convert-roundtrip"`.
2. Set `req.AgentEventsJSON` with the input events as a JSON array.
3. Root `Run` calls `crush_types.ToCrush` then `crush_types.FromCrush`, marshals the result events to JSON, and sets `resp.Output`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "convert-roundtrip"
	return nil
}
```
