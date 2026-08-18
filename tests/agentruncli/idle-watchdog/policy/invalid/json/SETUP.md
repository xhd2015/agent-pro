# Scenario

**Feature**: truncated JSON is a Read error

```
write `{` at IdlePolicyPath -> ReadIdlePolicy error
```

## Steps

1. Seed raw `{` (not via WriteIdlePolicy).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.RawFile = []byte("{")
	req.SessionID = "sess-policy-bad-json"
	return nil
}
```
