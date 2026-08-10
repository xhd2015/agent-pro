# Scenario

**Feature**: `NotifyTTYStarted` best-effort publish of `agent.tty.started`

```
NotifyTTYStarted(opts, sessionID, runner, workspace)
  -> empty URL: no HTTP
  -> URL set: Event type/source/payload; optional Bearer
  -> publish fail: warning: on WarnWriter; no error to caller
```

## Preconditions

- Injectable `EventBusOpts.Publisher` preferred (recordingPublisher in harness).
- Default session identity from root Setup.

## Steps

1. Grouping sets `Op=notify` and inject publisher defaults.
2. Leaf adjusts URL / token / failure mode.
3. Assert HTTP capture and/or WarnOutput.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opNotify
	req.UseInjectPublisher = true
	if req.Capture == nil {
		req.Capture = &HTTPCapture{}
	}
	return nil
}
```
