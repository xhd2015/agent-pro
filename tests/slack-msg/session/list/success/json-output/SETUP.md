# Scenario

**Feature**: session list --json

```
seed two map entries -> session list --json
  -> {"sessions":[...]} sorted updated_at desc; session_id + agent_session_id
```

## Steps

1. Seed fixture entries.
2. Args: session list --json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, sessionListFixtureEntries()); err != nil {
		return err
	}
	req.Args = []string{"session", "list", "--json"}
	return nil
}
```
