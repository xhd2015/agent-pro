# Scenario

**Feature**: nested scenario topic is embedded and showable

```
agent-pro skill verify-on-behalf-of-user scenario --show
-> depth labels, browser-agent, FAIL rules
```

## Steps

1. Invoke `agent-pro skill verify-on-behalf-of-user scenario --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "verify-on-behalf-of-user", "scenario", "--show"}
	return nil
}
```
