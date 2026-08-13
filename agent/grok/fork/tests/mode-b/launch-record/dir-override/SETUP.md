# Scenario

**Feature**: Mode B `--dir` overrides the grok working directory

```
fork.Main(["--session-id", id, "--dir", override])
  -> RunForeground dir = override
```

## Steps

1. Create override workspace.
2. Args include `--dir`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	override := filepath.Join(req.TempDir, "ws-override")
	if err := os.MkdirAll(override, 0o755); err != nil {
		return err
	}
	req.OverrideDir = absPath(t, override)
	req.Args = []string{"--session-id", fixtureSessionID, "--dir", req.OverrideDir}
	return nil
}
```
