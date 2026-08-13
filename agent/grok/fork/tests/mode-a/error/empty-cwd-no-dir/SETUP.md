# Scenario

**Feature**: session cwd empty and no `--dir` → error pass --dir

```
summary.json info.cwd = ""
fork.Main([]) -> error contains "pass --dir"
```

## Steps

1. Re-seed the fixture session with empty cwd (same uuid so Lsof still hits).

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Drop the workspace-keyed session from parent seed; write empty-cwd fixture.
	_ = os.RemoveAll(filepath.Join(req.GrokHome, "sessions"))
	dir := seedSession(t, req.GrokHome, fixtureSessionID, "")
	req.OpenFiles = map[int][]string{
		pidGrok: {lsofPath(dir)},
	}
	req.Args = []string{}
	return nil
}
```
