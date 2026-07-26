# Scenario

**Bug**: web grok-tty sync via `agentsync` — no duplicates (I1), no empty chat (I2), sync on open (I3)

```
I1: POST session (hello?) -> turn 1 done -> follow-up within overlap -> one user per prompt
I2: POST create (prompt only) -> delayed grok mock -> events.jsonl user + assistant
I3: seed finished + empty events + grok on disk -> GET session detail -> events.jsonl populated
```

## Preconditions

- Mock harness: llm-mock chrome + synthetic updates.jsonl scheduler.
- Assertion at `events.jsonl` file level (not Playwright).

## Steps

1. Grouping leaves configure web mode and follow-up timing.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode == "" {
		req.Mode = "web-rapid-followups"
	}
	return nil
}
```
