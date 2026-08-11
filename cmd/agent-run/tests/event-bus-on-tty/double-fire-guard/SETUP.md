# Scenario

**Feature**: ForceNew NotifyOnOpenPath + library WireOnTTYStarted → at most one publish

```
shared AlreadyNotified + URL set
  NotifyOnOpenPath("new-terminal")
  WireOnTTYStarted(opts)(TTYStartedInfo)
  -> PublishCount == 1
```

## Preconditions

- Models one open where both ForceNew label path and library true-TTY path would
  notify; production must share once-guard (`EventBusOpts.AlreadyNotified`).
- Injectable PublishHook; no real iTerm.

## Steps

1. Grouping sets `Op=double-fire`, inject publisher, UseAlreadyNotified.
2. Leaf sets URL and session identity.
3. Run calls both paths with shared opts.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opDoubleFire
	req.UseInjectPublisher = true
	req.UseAlreadyNotified = true
	if req.Capture == nil {
		req.Capture = &HTTPCapture{}
	}
	if req.EventBusURL == "" {
		req.EventBusURL = "http://127.0.0.1:23893"
	}
	return nil
}
```
