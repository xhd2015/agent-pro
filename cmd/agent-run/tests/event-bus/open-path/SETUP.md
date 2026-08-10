# Scenario

**Feature**: open-path publish policy — ForceNew notifies once; send never

```
# new-terminal (ForceNew success)
AppendEventBusFlags(follow-up) + NotifyOnOpenPath("new-terminal", …)
  -> PublishCount=1; ResultArgs include event-bus flags

# send (live)
NotifyOnOpenPath("send", …)
  -> PublishCount=0
```

## Preconditions

- Both leaves call `agentruncli.NotifyOnOpenPath` with the leaf OpenKind.
- Injectable recording Publisher; no real iTerm.
- Production: call with `"new-terminal"` only after successful ForceNew open.

## Steps

1. Grouping sets `Op=open-path` and inject publisher.
2. Leaf sets `OpenKind` (`new-terminal` | `send`) and URL.
3. Assert publish count and follow-up flags (new-terminal only).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opOpenPath
	req.UseInjectPublisher = true
	if req.Capture == nil {
		req.Capture = &HTTPCapture{}
	}
	if req.EventBusURL == "" {
		req.EventBusURL = "http://127.0.0.1:23891"
	}
	return nil
}
```
