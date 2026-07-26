# Scenario

**Feature**: `run --dir` pointing at a file (not a directory) exits non-zero

```
agent-run run --dir <TempDir/not-a-dir.txt> --agent-runner fake-codex "hi"
  -> exit ≠ 0
  -> stderr mentions not a directory
```

## Steps

1. Create a regular file under TempDir.
2. Pass that file path as `--dir`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	filePath := filepath.Join(req.TempDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("file\n"), 0o644); err != nil {
		return err
	}
	req.Args = append(req.Args, "--dir", filePath, "hi")
	return nil
}
```
