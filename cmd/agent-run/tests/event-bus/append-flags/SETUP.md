# Scenario

**Feature**: `AppendEventBusFlags` pure argv helper for ForceNew follow-up

```
AppendEventBusFlags(args, url, token)
  -> empty URL: unchanged
  -> URL set: append --event-bus-url and optional --event-bus-token
```

## Preconditions

- Pure function; no HTTP; no process env.

## Steps

1. Grouping sets `Op=append-flags`.
2. Leaf sets BaseArgs, URL, token.
3. Assert ResultArgs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opAppendFlags
	if len(req.BaseArgs) == 0 {
		req.BaseArgs = []string{"run", "--auto-send-or-resume", "--session-id", "s1"}
	}
	return nil
}
```
