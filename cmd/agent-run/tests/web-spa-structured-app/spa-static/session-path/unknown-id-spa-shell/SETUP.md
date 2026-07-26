# Scenario

**Feature**: unknown session id still serves SPA shell (deep link shell)

```
GET /sessions/spa-unknown-id-xyz -> 200 HTML with #root, no bootstrap script
```

## Preconditions

- No session directory for `spa-unknown-id-xyz` under home.

## Steps

1. Start web without seeding that id.
2. `GET /sessions/spa-unknown-id-xyz`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "session-path-unknown"
	req.SessionID = "spa-unknown-id-xyz"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPPath = "/sessions/" + req.SessionID
	req.HTTPAuth = "none"
	return nil
}
```
