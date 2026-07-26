# Scenario

**Feature**: `run` without `--dir` defaults workspace to process cwd

```
cwd=TempDir
agent-run run --json --agent-runner fake-codex "hi"
  -> exit 0
  -> meta.workspace = abs(TempDir)  # process cwd
```

## Steps

1. Create an unused sibling directory (must not become workspace).
2. Run without `--dir`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Sibling must not be chosen when --dir is omitted.
	other := filepath.Join(req.TempDir, "other-ws")
	if err := os.MkdirAll(other, 0o755); err != nil {
		return err
	}
	req.Args = append(req.Args, "--json", "hi")
	return nil
}
```
