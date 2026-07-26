# Scenario

**Feature**: print session with meta only — no events file yet

```
CreateSession(finished) without AppendEvent -> --print -> header 0 lines + (no events yet)
```

## Preconditions

- `meta.json` exists; `events.jsonl` is absent.

## Steps

1. Seed `print_no_events` with status `finished` only.
2. Run print for that bare session id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const noEventsSessionID = "print_no_events"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, noEventsSessionID, "finished")
	req.SessionID = noEventsSessionID
	req.Args = printSessionArgs(noEventsSessionID)
	return nil
}
```
