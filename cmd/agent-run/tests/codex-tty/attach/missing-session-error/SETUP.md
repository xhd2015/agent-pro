# Scenario

**Feature**: attach with unknown or expired session id fails clearly

```
agent-run attach session-999999 → exit 1, stderr mentions not found or expired
```

## Steps

1. Run `agent-run attach session-999999` with empty registry.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"attach", "session-999999"}
	return nil
}
```