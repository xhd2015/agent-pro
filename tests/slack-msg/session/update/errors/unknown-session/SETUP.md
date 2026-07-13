# Scenario

**Feature**: session update unknown session

```
--session-id not-in-map --dir PATH -> session not found; exit 1
```

## Steps

1. Empty map; real directory for --dir.
2. Unknown session id.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{}); err != nil {
		return err
	}
	dir := filepath.Join(req.WorkDir, "ws-unknown")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	req.Args = []string{
		"session", "update",
		"--session-id", "slack-unknown-upd",
		"--dir", dir,
	}
	return nil
}
```
