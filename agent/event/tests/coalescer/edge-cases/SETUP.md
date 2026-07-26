# Scenario

**Feature**: Grouping node for edge cases involving empty text and boundary conditions

## Preconditions
- Grouping node for edge cases involving empty text and boundary conditions.
- Coalescer behavior must be consistent regardless of text content.

## Steps
1. Test sequences with empty text on various phases.
2. Verify coalescer rules still apply: show phases mark ID as shown, PhaseEnd gets skipped.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Edge cases: empty text across phases.
	t.Helper()
	return nil
}
```
