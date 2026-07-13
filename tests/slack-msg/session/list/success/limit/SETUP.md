# Scenario

**Feature**: session list --limit after sort

```
two entries -> session list --limit 1
  -> only the newest row (plus human header)
```

## Steps

1. Seed two fixture entries.
2. Args: session list --limit 1.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, sessionListFixtureEntries()); err != nil {
		return err
	}
	req.Args = []string{"session", "list", "--limit", "1"}
	return nil
}
```
