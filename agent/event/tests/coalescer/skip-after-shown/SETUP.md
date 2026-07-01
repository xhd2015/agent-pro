# Scenario

**Feature**: Grouping node for sequences where a `PhaseEnd` follows a "show" phase (start/update/instant)

## Preconditions
- Grouping node for sequences where a `PhaseEnd` follows a "show" phase (start/update/instant).
- All events in a leaf share the same `ID`.
- The first event marks the ID as "shown", so the trailing `PhaseEnd` must be skipped.

## Steps
1. Feed a sequence where the first event is a show-phase, second is PhaseEnd.
2. The show-phase must not be skipped; the PhaseEnd must be skipped.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Leaves in this group: first event is a show-phase, second is PhaseEnd (same ID).
	// No shared field overrides needed — each leaf sets its own Events.
	t.Helper()
	return nil
}
```
