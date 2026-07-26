# Scenario

**Feature**: PUT missing path returns 400 and does not change config (A3)

```
PUT /workspace {"path":"/…/does-not-exist"}
  -> 400
  -> config selected_workspace still baseline
```

## Preconditions

- Baseline selected path from parent invalid/ SETUP.
- Missing absolute path under TempDir.

## Steps

1. Start web; PUT non-existent path.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "missing-path-400"
	missing := filepath.Join(req.TempDir, "workspaces", "does-not-exist")
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "put", Method: "PUT", Path: "/api/agent-run/workspace", Body: putWorkspaceBody(missing)},
	}
	return nil
}
```
