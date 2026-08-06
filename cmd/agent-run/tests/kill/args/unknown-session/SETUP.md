# Scenario

**Feature**: `kill` of a session id with no registry entry fails

```
empty AGENT_RUN_HOME -> agent-run kill no-such-id -> exit 1; not found / expired
```

## Steps

1. Args = `["kill", "no-such-kill-session"]` with empty home (default).
2. Run Mode handle.
3. Assert exit non-zero and not-found wording.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{"kill", "no-such-kill-session"}
	return nil
}
```
