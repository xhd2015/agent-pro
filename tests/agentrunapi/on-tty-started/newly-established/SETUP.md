# Scenario

**Feature**: first open establishes TTY — OnTTYStarted policy for ModeRun

```
AutoSendOrResume(missing session → ModeRun, RunSession hook)
  + OnTTYStarted set|nil
  -> fires once | no-op
```

## Preconditions

- Session not seeded → Classify ModeRun.
- `RunSession` hook installed (no real agentui / binary).
- Leaves set `InstallHook` true or false.

## Steps

1. Grouping sets `Op=newly-established`.
2. Leaf sets whether `OnTTYStarted` is installed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opNewlyEstablished
	return nil
}
```
