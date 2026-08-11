# Scenario

**Feature**: live follow-up into existing TTY does not re-fire OnTTYStarted

```
AutoSendOrResume(ModeRun)  -> OnTTYStarted once
AutoSendOrResume(ModeSend) -> OnTTYStarted not called again
```

## Preconditions

- Same recording OnTTYStarted across both AutoSendOrResume calls.
- Second call: seeded live session + probe `RunnerExited=false` → ModeSend.
- `SendLive` hook only (no real agentsend).

## Steps

1. Grouping sets `Op=follow-up-live`.
2. Leaf sets session identity.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opFollowUpLive
	req.InstallHook = true
	return nil
}
```
