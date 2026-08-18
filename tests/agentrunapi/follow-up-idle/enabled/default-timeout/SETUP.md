# Scenario

**Feature**: enabled + timeout 0 defaults to `10m`

```
BuildFollowUpCommand(ExitOnIdle:true, IdleTimeout:0, Open, SessionID, Prompt)
  -> tokens --exit-on-idle and --idle-timeout=10m before --
```

## Steps

1. Set `IdleTimeout=0` with `ExitOnIdle=true`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = 0
	req.SessionID = "sess-idle-on-default"
	req.Prompt = "idle on default"
	return nil
}
```
