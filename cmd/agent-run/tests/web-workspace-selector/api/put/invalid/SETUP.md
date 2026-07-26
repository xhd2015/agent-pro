# Scenario

**Feature**: PUT rejects non-directory / missing paths with 400 (A3)

```
# invalid path kinds
PUT missing path | PUT existing file
  -> 400
  -> selected_workspace / recent unchanged
```

## Preconditions

- Baseline config optionally seeded with a known selected path to prove no clobber.
- Leaves choose missing vs file.

## Steps

1. Seed optional baseline selected_workspace.
2. PUT invalid path; assert 400 + config unchanged.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Baseline selected path so we can prove invalid PUT does not clobber it.
	baseline := makeSelectDir(t, req, "baseline-ws")
	req.SelectPath = baseline // reused as "should remain" for invalid leaves
	if err := writeHomeConfig(t, req.Home, map[string]any{
		"selected_workspace": baseline,
		"recent_workspaces":  []string{baseline},
	}); err != nil {
		return err
	}
	// WebWorkingDir under TempDir so process is sandboxed.
	req.WebWorkingDir = mustMkdir(t, filepath.Join(req.TempDir, "proc-cwd"))
	return nil
}
```
