# Scenario

**Feature**: nested tty topic is embedded and showable

```
agent-pro skill verify-on-behalf-of-user tty --show
-> tty-watch, run --detach, kill reclaim
```

## Steps

1. Invoke `agent-pro skill verify-on-behalf-of-user tty --show`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "verify-on-behalf-of-user", "tty", "--show"}
	return nil
}
```
