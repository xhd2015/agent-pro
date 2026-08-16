# Scenario

**Feature**: session update --json returns full updated map entry

```
session update --session-id ID --dir PATH --json
  -> JSON object with session_id + agent_session_id + dir + preserved fields
```

## Steps

1. Create workspace directory.
2. Args include --json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	abs, err := ensureUpdateWorkspaceDir(t, req)
	if err != nil {
		return err
	}
	req.Args = []string{
		"session", "update",
		"--session-id", sessionUpdateFixtureID,
		"--dir", abs,
		"--json",
	}
	return nil
}
```
