# Scenario

**Feature**: fresh install has no selected workspace; status.workspace ≈ process cwd (A1)

```
# empty config → status.workspace equals process cwd
WebWorkingDir=/…/proc-cwd; no selected_workspace
  -> GET /api/agent-run/status
  -> workspace == process_cwd == WebWorkingDir
  -> recent_workspaces empty; process_cwd + home present
```

## Preconditions

- No `config.json` selected_workspace (fresh home).
- Process started with `cmd.Dir = WebWorkingDir`.

## Steps

1. Create process cwd dir; set `WebWorkingDir`.
2. Start web; `GET /api/agent-run/status`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "fresh-defaults-to-cwd"
	cwd := mustMkdir(t, filepath.Join(req.TempDir, "proc-cwd"))
	req.WebWorkingDir = cwd
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "status", Method: "GET", Path: "/api/agent-run/status"},
	}
	return nil
}
```
