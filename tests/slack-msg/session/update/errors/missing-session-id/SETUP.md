# Scenario

**Feature**: session update without session id

```
session update --dir PATH (no session id) -> session id required; exit 1
```

## Steps

1. Create a real dir so only id is missing.
2. Args without --session-id / env.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkDir, "ws-missing-id")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	req.Args = []string{"session", "update", "--dir", dir}
	return nil
}
```
