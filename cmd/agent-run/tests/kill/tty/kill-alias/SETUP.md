# Scenario

**Feature**: `tty kill <session-id>` stops a live session like top-level `kill`

```
startLiveKillFixture(kill-tty-alias-1)
agent-run tty kill kill-tty-alias-1
  -> stdout: stopped kill-tty-alias-1\n
  -> process dead; registry removed
```

## Steps

1. Start live fixture with meta.
2. Args = `tty kill kill-tty-alias-1`.
3. Assert stopped success path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	const sid = "kill-tty-alias-1"
	startLiveKillFixture(t, req, sid, true)
	req.Mode = "handle"
	req.Args = []string{"tty", "kill", sid}
	return nil
}
```
