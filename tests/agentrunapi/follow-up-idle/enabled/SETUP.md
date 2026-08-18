# Scenario

**Feature**: `ExitOnIdle=true` enables idle-exit emit and timeout validation

```
NormalizeIdle(true, timeout) -> enabled + default/explicit duration, or error if negative
BuildFollowUpCommand -> --exit-on-idle and --idle-timeout=<compact> before --
```

## Steps

1. Grouping sets `ExitOnIdle=true`.
2. Leaves vary `IdleTimeout` (0 → 10m, `2m`, `2s`, `-1s`).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ExitOnIdle = true
	return nil
}
```
