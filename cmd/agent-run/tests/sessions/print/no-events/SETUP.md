# Scenario

**Feature**: print session with meta only — no events file yet

```
CreateSession(finished) without AppendEvent -> --print -> header 0 lines + (no events yet)
```

## Preconditions

- `meta.json` exists; `events.jsonl` is absent.

## Steps

1. Seed `fake-codex/print_no_events` with status `finished` only.
2. Run print for that session.

```go
import "testing"

const noEventsSessionID = "print_no_events"

func Setup(t *testing.T, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, printRunner, noEventsSessionID, "finished")
	req.SessionID = noEventsSessionID
	req.Args = printSessionArgs(printRunner, noEventsSessionID)
	return nil
}
```