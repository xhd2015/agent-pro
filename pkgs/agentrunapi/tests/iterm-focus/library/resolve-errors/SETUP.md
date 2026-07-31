# Scenario

**Feature**: FocusSession fails before tree/iTerm when session is unusable

```
missing session_id | unknown session -> error; FocusITerm never called
```

## Steps

1. Leaves omit or seed no matching session.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Home exists; no successful resolve.
	req.SeedSession = false
	req.SeedRegistry = false
	return nil
}
```
