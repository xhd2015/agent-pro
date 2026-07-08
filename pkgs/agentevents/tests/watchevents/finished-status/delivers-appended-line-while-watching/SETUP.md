# Scenario

**Bug**: C1 — `WatchEvents` delivers line appended after session is `finished`

```
seed finished grok-tty/chat_tail_watch
  -> WatchEvents from EOF offset
  -> append WATCHEVENTS_FINISHED_APPEND_MARKER
```

## Steps

1. Set session id and append marker.
2. Default timing from root `Setup`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "chat_tail_watch"
	req.AppendText = defaultWatchAppendMarker
	return nil
}
```