# Scenario

**Feature**: `ExitOnIdle=false` is a silent no-op (timeout ignored)

```
NormalizeIdle(false, timeout) -> enabled=false, no error
BuildFollowUpCommand -> no --exit-on-idle / --idle-timeout tokens
```

## Steps

1. Grouping sets `ExitOnIdle=false`.
2. Leaves vary `IdleTimeout` (0, `2s`, `-1s`).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ExitOnIdle = false
	return nil
}
```
