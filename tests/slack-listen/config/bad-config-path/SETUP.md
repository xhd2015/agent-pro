# Scenario

**Feature**: bad config path fails fast

```
slack-listen listen --config missing.json -> failed to load config
```

## Steps

1. Pass non-existent config path with tokens.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WorkDir = t.TempDir()
	missing := filepath.Join(req.WorkDir, "missing-slack-config.json")
	req.Args = []string{"--config", missing}
	return nil
}
```