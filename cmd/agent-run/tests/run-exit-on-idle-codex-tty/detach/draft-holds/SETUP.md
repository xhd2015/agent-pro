# Scenario

**Feature**: `--exit-on-idle` holds Codex when a real no-submit draft occupies the composer

```
detach --exit-on-idle --idle-timeout=10s
  -> wait sendable / policy written
  -> tty send --no-submit DRAFT_OCCUPANCY_HOLD_zz9
  -> wait 10s + 5s grace + 20s probe slack
  -> tty still live
```

Space probe must not collapse a real draft. Do not submit the draft (that would
start a turn and make the screen busy).

## Steps

1. Fixed session id + distinctive draft; observe after the same timeout+grace+slack.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "draft-hold"
	req.SessionID = "idle-codex-draft-hold-1"
	req.Draft = defaultDraft
	req.ObserveAfter = defaultIdleTimeout + defaultGrace + probeSlack
	return nil
}
```
