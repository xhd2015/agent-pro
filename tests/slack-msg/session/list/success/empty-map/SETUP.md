# Scenario

**Feature**: session list with empty map

```
empty sessions.json -> session list -> empty stdout; exit 0
```

## Steps

1. Seed empty map.
2. Args: session list.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{}); err != nil {
		return err
	}
	req.Args = []string{"session", "list"}
	return nil
}
```
