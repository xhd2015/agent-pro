# Scenario

**Feature**: browser UI renders CLI-parity timeline and terminal modal attach

```
chat page -> AgentEvent rows (no phased coalescing)
terminal modal -> attach relay backend
```

## Preconditions

- `playwright-debug` on PATH for UI leaves.
- Web server running with explicit token.

## Steps

1. Grouping setup sets `req.Area = "frontend-ui"` and `req.Mode = "ui"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Area = "frontend-ui"
	req.Mode = "ui"
	return nil
}
```
