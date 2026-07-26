# Scenario

**Feature**: `GET /api/agent-run/status` workspace resolution and persistence

```
# status reflects selected_workspace or falls back to process cwd
GET /api/agent-run/status -> workspace, process_cwd, home, recent_workspaces
```

## Preconditions

- Known process cwd via `WebWorkingDir` when asserting default workspace.
- No pre-seeded `selected_workspace` for fresh default; PUT then GET for persist.

## Steps

1. Leaves set cwd and/or PUT steps, then probe status.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Status group: leaves compose GET /status (and optional PUT) steps.
	if req.Scenario == "" {
		req.Scenario = "status"
	}
	return nil
}
```
