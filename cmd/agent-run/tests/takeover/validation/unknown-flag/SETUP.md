# Scenario

**Feature**: unknown flag on `takeover` fails cleanly

```
agent-run takeover --not-a-real-takeover-flag <provider-session-id>
  -> exit non-zero
  -> unknown / unrecognized flag (not crash; not silent success)
```

## Steps

1. Pass a junk flag with a session id so the error is flag-related.
2. Run Mode handle.
3. Assert non-zero exit and unknown-flag wording (command must be registered).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{
		"takeover",
		"--not-a-real-takeover-flag",
		takeoverFixtureSessionID,
	}
	return nil
}
```
