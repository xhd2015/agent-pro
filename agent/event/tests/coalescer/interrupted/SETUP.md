## Preconditions
- Grouping node for sequences where a non-`ActionMessage` event interrupts message flow.
- Non-`ActionMessage` events always pass through and reset the coalescer state.
- After reset, the next `ActionMessage` is treated as if nothing preceded it.

## Steps
1. Insert a non-`ActionMessage` event (like `ActionError` or `ActionToolCall`) between PhaseEnd events.
2. The non-message event resets state; subsequent PhaseEnd with same ID is not skipped.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Interruption tests: non-ActionMessage events reset coalescer state.
	t.Helper()
	return nil
}
```
