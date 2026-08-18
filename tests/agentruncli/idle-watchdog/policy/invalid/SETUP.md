# Scenario

**Feature**: invalid policy file → ReadIdlePolicy error

```
raw `{` or idle_timeout=nope -> ReadIdlePolicy error (found optional)
```

## Steps

1. Grouping documents invalid seed (raw file).
2. Leaves set RawFile.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opPolicy
	req.WritePolicy = false
	return nil
}
```
