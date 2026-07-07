# Scenario

**Feature**: successful grok bind emits resolve progress think event

```
mock hook seeds updates.jsonl
  -> DiscoverSession succeeds
  -> events.jsonl contains think "Resolve session id..."
```

## Steps

1. Configure success binding env (mock seeds `updates.jsonl`).
2. Start web and POST `grok-tty` session.
3. `Run` waits for `finished` and reads `events.jsonl`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "binding-progress-emitted"
	configureBindingSuccessEnv(t, req, "resolve progress probe", enhanceChatSuccessMarker)
	startWebGrokSession(t, req)
	return nil
}
```