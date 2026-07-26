# Scenario

**Feature**: auth.test returns error

```
slack-msg auth status -> auth.test invalid_auth -> auth failed:
```

## Steps

1. Use auth-fail slacktest server.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AuthAPIFail = true
	req.Args = []string{
		"auth", "status",
		"--token", slackTestToken,
	}
	return nil
}
```
