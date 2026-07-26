# Scenario

**Feature**: resume (and auto→resume) spawn the provider in `meta.workspace`
unless `--dir` overrides; missing `meta.workspace` errors with a `--dir` hint

```
meta.workspace = /created/ws; CLI cwd ≠ that
  -> resume / auto → child cwd = meta.workspace
  -> resume --dir OVERRIDE → child cwd = OVERRIDE

meta.workspace = /gone-ws (missing); no --dir
  -> resume / auto → exit 1; session workspace missing + path + --dir hint
```

## Steps

1. Create distinct created-ws and cli-cwd directories under TempDir.
2. Seed exited meta with Workspace = created-ws (missing-workspace leaves override
   Workspace to a never-created path under TempDir).
3. Set `req.WorkDir` = cli-cwd so process cwd differs.
4. Install argv+cwd recording runner (happy-path leaves; error leaves still may
   install for consistency).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Shared layout helpers used by workspace leaves.
	created := filepath.Join(req.TempDir, "created-ws")
	cliCwd := filepath.Join(req.TempDir, "cli-cwd")
	for _, d := range []string{created, cliCwd} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
		// Project-like marker (avoid bare-tmp heuristics if any).
		if err := os.WriteFile(filepath.Join(d, "README.md"), []byte("workspace leaf\n"), 0644); err != nil {
			return err
		}
	}
	req.Workspace = created
	req.WorkDir = cliCwd
	return nil
}
```
