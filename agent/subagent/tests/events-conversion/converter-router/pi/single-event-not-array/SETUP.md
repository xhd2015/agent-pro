## Preconditions
- **After fix**: `ConvertRawLine` for pi must handle both JSON arrays `[{...}]` and single objects `{...}`.
- The pi runner writes SINGLE JSON objects (one per line), not arrays.
- After the fix, `json.Unmarshal` into a single `pi_types.Event` must succeed when array unmarshal fails.
- This test verifies that a single pi event (not wrapped in array) is accepted without error.

## Steps
1. Set `req.AgentRunner = "pi"`.
2. Feed a single pi message_start event (NOT wrapped in array).
3. Verify ConvertRawLine returns no error and produces correct AgentEvent output.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.AgentRunner = "pi"
    req.RawJSON = `{"type":"message_start","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}`
    return nil
}
```
