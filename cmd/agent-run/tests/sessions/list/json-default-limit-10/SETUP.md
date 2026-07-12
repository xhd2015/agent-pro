# Scenario

**Feature**: `sessions --json` uses same default limit of 10

```
seed 15 -> sessions --json -> sessions array length 10, sorted desc by updated_at
```

## Preconditions

- Q3: JSON uses same default limit and `--limit`.

## Steps

1. Seed 15 sessions.
2. Run `agent-run sessions --json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	seedNSessions(t, req.Home, 15)
	req.Args = append(req.Args, "--json")
	return nil
}
```
