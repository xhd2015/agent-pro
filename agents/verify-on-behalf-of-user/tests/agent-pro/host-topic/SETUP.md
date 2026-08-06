# Scenario

**Feature**: nested host topic is embedded and showable

```
agent-pro skill verify-on-behalf-of-user host --show
-> opt-in host mode, wrk ladder, dry-run, change-scoped targets
```

## Steps

1. Invoke `agent-pro skill verify-on-behalf-of-user host --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "verify-on-behalf-of-user", "host", "--show"}
	return nil
}
```
