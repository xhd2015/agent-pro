# Scenario

**Feature**: kill a live keep-alive fixture → stopped; session no longer live

```
startLiveKillFixture(kill-live-1)
agent-run kill kill-live-1
  -> stdout: stopped kill-live-1\n
  -> fixture PID dead; registry removed
```

## Steps

1. Start live fixture with meta.
2. Args = `kill kill-live-1`.
3. Assert stopped line + idle (process dead, registry gone).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	const sid = "kill-live-1"
	startLiveKillFixture(t, req, sid, true)
	req.Mode = "handle"
	req.Args = []string{"kill", sid}
	return nil
}
```
