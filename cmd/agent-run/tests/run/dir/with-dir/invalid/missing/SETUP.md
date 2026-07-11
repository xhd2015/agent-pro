# Scenario

**Feature**: `run --dir` with a non-existent path exits non-zero

```
agent-run run --dir <TempDir/no-such-dir> --agent-runner fake-codex "hi"
  -> exit ≠ 0
  -> stderr mentions missing / not found / does not exist
```

## Steps

1. Point `--dir` at a path under TempDir that was never created.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	missing := filepath.Join(req.TempDir, "no-such-dir")
	req.Args = append(req.Args, "--dir", missing, "hi")
	return nil
}
```
