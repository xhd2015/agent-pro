# Scenario

**Feature**: listen bad config path fails fast

```
slack-msg listen --config missing.json -> failed to load config
```

## Steps

1. Pass non-existent config path with tokens.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WorkDir = t.TempDir()
	missing := filepath.Join(req.WorkDir, "missing-slack-config.json")
	req.Args = []string{"--config", missing}
	return nil
}
```
