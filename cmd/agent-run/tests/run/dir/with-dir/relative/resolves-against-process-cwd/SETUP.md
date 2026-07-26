# Scenario

**Feature**: `run --dir <relative>` resolves against process cwd to an absolute workspace

```
cwd=TempDir
agent-run run --dir rel-ws --json --agent-runner fake-codex "hi"
  -> meta.workspace = abs(TempDir/rel-ws)
```

## Steps

1. Create `rel-ws` under TempDir (process cwd).
2. Pass relative `--dir rel-ws`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	rel := "rel-ws"
	ws := filepath.Join(req.TempDir, rel)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	req.Args = append(req.Args, "--dir", rel, "--json", "hi")
	return nil
}
```
