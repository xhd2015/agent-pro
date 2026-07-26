# Scenario

**Feature**: Grouping node for consecutive `PhaseEnd` handling

## Preconditions
- Grouping node for consecutive `PhaseEnd` handling.
- Tests how the coalescer treats back-to-back PhaseEnd events with same or different IDs.

## Steps
1. Feed sequences of PhaseEnd events.
2. Verify same-ID duplicates are skipped; different-ID pairs are not.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Duplicate/consecutive PhaseEnd tests.
	t.Helper()
	return nil
}
```
