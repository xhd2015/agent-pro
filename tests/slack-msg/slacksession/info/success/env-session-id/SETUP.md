# Scenario

**Feature**: session info resolves session id from env

```
SLACK_MSG_SESSION_ID=ID (no --session-id) -> session info succeeds
```

## Steps

1. Set env only; args: session info.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Env = append(req.Env, "SLACK_MSG_SESSION_ID="+sessionInfoFixtureID)
	req.Args = []string{"session", "info"}
	return nil
}
```
