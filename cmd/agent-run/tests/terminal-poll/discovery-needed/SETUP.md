# Scenario

**Bug**: terminal discovery poll must stop after available:true

```
running tty session without terminal_session_id in detail
  -> slow /terminal discovery until registry live
  -> stop polling once available:true
```

## Preconditions

- Session detail omits `terminal_session_id` at load.
- PTY registry may appear after a short delay.

## Steps

1. Seed running session without terminal mapping in meta.
2. Delay registry write until discovery window begins.
3. Monitor terminal GET traffic; assert bounded polls that stop after available.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "discovery-needed"
	return nil
}
```