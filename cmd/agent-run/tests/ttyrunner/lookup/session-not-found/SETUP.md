# Scenario

**Feature**: unknown terminal session id returns error

```
LookupSession(session-missing) -> not found error
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "session-not-found"
	return nil
}
```
