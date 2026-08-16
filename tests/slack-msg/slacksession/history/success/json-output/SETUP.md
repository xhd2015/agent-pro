# Scenario

**Feature**: session history --json

```
session history --session-id ID --json -> JSON document chronological
```

## Steps

1. Pass --json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"session", "history",
		"--session-id", sessionHistoryFixtureID,
		"--json",
	}
	return nil
}
```
