# Scenario

**Feature**: Grouping node for standalone single-event tests

## Preconditions
- Grouping node for standalone single-event tests.
- Each event is fed in isolation; no prior state in the coalescer.

## Steps
1. Feed a single event to the coalescer.
2. Verify it is never skipped (standalone events always display).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Standalone tests: each leaf defines exactly one event.
	// No shared setup needed beyond root.
	t.Helper()
	return nil
}
```
