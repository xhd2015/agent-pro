# Scenario

**Feature**: session history human lines oldest→newest

```
session history --session-id ID -> three human lines chronological
```

## Steps

1. Default human output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"session", "history",
		"--session-id", sessionHistoryFixtureID,
	}
	return nil
}
```
