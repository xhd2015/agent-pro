# Scenario

**Feature**: session info human key: value output

```
session info --session-id ID -> key: value lines incl. message_count, session_dir
```

## Steps

1. Args: session info --session-id fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"session", "info",
		"--session-id", sessionInfoFixtureID,
	}
	return nil
}
```
