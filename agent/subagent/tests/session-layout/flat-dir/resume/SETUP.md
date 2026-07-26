# Scenario

**Feature**: flat session resume appends to existing artifacts

```
# same SessionID + SessionLayout on second Run
first Run -> events; second Run -> append without clobbering host meta
```

## Preconditions

- Flat session dir with pre-created host-owned `meta.json`.

## Steps

1. Descendant leaves pre-create meta and configure two mock configs.

## Context

- Resume uses `agent_session_id` alias for session matching.

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
