# Scenario

**Feature**: session info --json

```
session info --session-id ID --json -> JSON object with message_count + session_dir
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
		"session", "info",
		"--session-id", sessionInfoFixtureID,
		"--json",
	}
	return nil
}
```
