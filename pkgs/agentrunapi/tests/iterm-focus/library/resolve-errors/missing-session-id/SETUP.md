# Scenario

**Feature**: empty session_id is rejected

```
FocusOpts{SessionID: ""} -> FocusSession error; no focus
```

## Steps

1. Leave SessionID empty (whitespace-only also invalid if implementer trims).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = ""
	req.SeedSession = false
	return nil
}
```
