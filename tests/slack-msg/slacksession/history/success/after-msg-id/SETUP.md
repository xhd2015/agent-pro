# Scenario

**Feature**: session history --after-msg-id filters to later messages

```
--after-msg-id m1 -> only m2 and m3 lines
```

## Steps

1. Pass --after-msg-id m1.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"session", "history",
		"--session-id", sessionHistoryFixtureID,
		"--after-msg-id", "m1",
	}
	return nil
}
```
