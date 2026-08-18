# Scenario

**Feature**: unparseable `idle_timeout` is a Read error

```
{"exit_on_idle":true,"idle_timeout":"nope"} -> ReadIdlePolicy error
```

## Steps

1. Seed valid JSON with `idle_timeout=nope`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.RawFile = []byte(`{"exit_on_idle":true,"idle_timeout":"nope"}`)
	req.SessionID = "sess-policy-bad-dur"
	return nil
}
```
