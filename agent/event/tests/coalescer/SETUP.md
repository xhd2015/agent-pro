# Scenario

**Feature**: The `agent/event/print` package provides a `Coalescer` struct

## Preconditions

- The `agent/event/print` package provides a `Coalescer` struct.
- `Coalescer` tracks the last `ActionMessage` ID and whether content has been shown.
- `Coalescer.ShouldSkip(event AgentEvent) bool` returns `true` when the event should be suppressed.

## Steps

1. Create a `Coalescer` instance.
2. Feed each event from `req.Events` sequentially through `ShouldSkip`.
3. Collect the result for each event as a boolean in `resp.Skipped`.

## Context

- `Request.Events` is the sequence of `AgentEvent` to feed.
- `Response.Skipped[i]` is `true` when `ShouldSkip` returned `true` for `Events[i]`.

```go
import (
	"testing"

	print "github.com/xhd2015/agent-pro/agent/event/print"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = assertAllDisplayed
	_ = assertSkipAt
	return nil
}

func assertAllDisplayed(t *testing.T, skipped []bool) {
	t.Helper()
	for i, s := range skipped {
		if s {
			t.Fatalf("event[%d] unexpectedly skipped", i)
		}
	}
}

func assertSkipAt(t *testing.T, skipped []bool, indexes ...int) {
	t.Helper()
	skipSet := make(map[int]bool)
	for _, i := range indexes {
		skipSet[i] = true
	}
	for i, s := range skipped {
		if skipSet[i] && !s {
			t.Fatalf("event[%d] should be skipped but was not", i)
		}
		if !skipSet[i] && s {
			t.Fatalf("event[%d] should not be skipped but was", i)
		}
	}
}
```
