# Scenario

**Feature**: traceSession honors SessionLayout path overrides

```
# Run writes events; traceSession(CatchUp) reads same EventsPath
subagent.Run -> traceSession -> formatted Events: N lines output
```

## Preconditions

- Flat session with custom or default `EventsPath`.

## Steps

1. Descendant leaves configure layout and optional custom events path.

## Context

- Uses `runTrace` helper from root `DOCTEST.md`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.AgentRunner == "" {
		req.AgentRunner = "fake-codex"
	}
	return nil
}```
