# Scenario

**Feature**: chat UI renders think/error event cards from SSE timeline

```
chat page -> progress-card for think events
chat page -> error-card for bind failure
terminal modal unchanged (separate from chat cards)
```

## Preconditions

- `playwright-debug` on PATH for UI leaves.
- Web server running with grok mock flags.

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