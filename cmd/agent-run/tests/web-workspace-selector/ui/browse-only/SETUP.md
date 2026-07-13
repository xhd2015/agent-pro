# Scenario

**Feature**: Quick chips and Recent only change browse path — never auto-commit

```
# critical product rule
tap Quick Home | Recent row
  -> browser path updates
  -> GET /status workspace UNCHANGED (no PUT)
```

## Preconditions

- Selector page mounted.
- Known selected_workspace different from chip/recent target so change would be visible.
- Leaves seed config and assert via page.evaluate fetch or dual check.

## Steps

1. Seed selected workspace ≠ Home / recent target.
2. Open selector; tap control; re-check status API workspace.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Shared: selected workspace is a known dir under TempDir (not $HOME).
	req.SelectPath = makeSelectDir(t, req, "already-selected")
	req.WebWorkingDir = mustMkdir(t, filepath.Join(req.TempDir, "proc-cwd"))
	if err := writeHomeConfig(t, req.Home, map[string]any{
		"selected_workspace": req.SelectPath,
		"recent_workspaces":  []string{req.SelectPath},
	}); err != nil {
		return err
	}
	return nil
}
```
