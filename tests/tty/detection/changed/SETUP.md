# Scenario

**Feature**: Changed raw byte equality (newlines count)

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opChanged
	return nil
}
```
