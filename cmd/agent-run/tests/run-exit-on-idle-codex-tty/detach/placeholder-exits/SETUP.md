# Scenario

**Feature**: `--exit-on-idle` reaps Codex after timeout on empty/placeholder composer

```
detach --exit-on-idle --idle-timeout=10s + finished mock turn
  -> placeholder/empty composer (space probe → empty)
  -> wait 10s + 5s grace + 20s probe slack
  -> tty not live
```

DetectInputBox may still say occupied (`default ·` / hint without glue). Occupancy
truth is the watchdog space probe, observed as session gone.

## Steps

1. Fixed session id; observe after timeout+grace+slack. Do not inject a draft.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "placeholder-exit"
	req.SessionID = "idle-codex-placeholder-1"
	req.ObserveAfter = defaultIdleTimeout + defaultGrace + probeSlack
	return nil
}
```
