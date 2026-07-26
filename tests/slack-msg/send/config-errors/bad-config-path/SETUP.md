# Scenario

**Feature**: missing --config file

```
Caller -> slack-msg send --config /missing.json MESSAGE -> failed to load config
```

## Steps

1. Use nonexistent absolute config path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	missing := filepath.Join(req.WorkDir, "does-not-exist.json")
	req.Args = []string{"send", "--config", missing, "Hello"}
	return nil
}
```
