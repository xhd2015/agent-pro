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
import "testing"

func Setup(t *testing.T, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, sessionListFixtureEntries()); err != nil {
		return err
	}
	req.Args = []string{"session", "list", "--json"}
	return nil
}
```
