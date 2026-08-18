# Scenario

**Feature**: idle-policy.json path, read/write, and policy-gated start

```
IdlePolicyPath(home, session) -> sessions/<id>/idle-policy.json
WriteIdlePolicy -> ReadIdlePolicy -> found / fields
missing or exit_on_idle=false -> NewIdleWatchdog Tick never exits
```

## Steps

1. Grouping sets `Op=policy`.
2. Leaves write, skip, or seed raw JSON; some then Tick.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opPolicy
	return nil
}
```
