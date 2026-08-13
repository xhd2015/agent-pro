# Scenario

**Feature**: Mode A `--dir` overrides the open directory and is forwarded

```
fork.Main(["--dir", override])
  -> OpenInNewTerminal(override, <exe> --session-id <id> --dir <override>)
```

## Steps

1. Create `ws-override` under temp.
2. Args `["--dir", override]`.

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
	req.Args = []string{"--dir", req.OverrideDir}
	return nil
}
```
