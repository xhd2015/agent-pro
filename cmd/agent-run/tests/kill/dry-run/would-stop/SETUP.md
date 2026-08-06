# Scenario

**Feature**: dry-run on a live fixture prints would-stop and does not end the session

```
startLiveKillFixture(kill-dry-1)
agent-run kill --dry-run kill-dry-1
  -> dry-run: would stop kill-dry-1\n
  -> sleep PID still alive; registry file remains
```

## Steps

1. Start live registry fixture (`kill-dry-1`) with session meta.
2. Args = `kill --dry-run kill-dry-1`.
3. Assert dry-run stdout; process and registry still present.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	const sid = "kill-dry-1"
	startLiveKillFixture(t, req, sid, true)
	req.Mode = "handle"
	req.Args = []string{"kill", "--dry-run", sid}
	return nil
}
```
