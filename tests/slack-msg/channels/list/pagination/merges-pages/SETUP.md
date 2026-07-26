# Scenario

**Feature**: cursor pages merged then sorted

```
slack-msg channels list --token -> same three lines as single-page fixture
```

## Steps

1. Flags for token only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
	}
	return nil
}
```
