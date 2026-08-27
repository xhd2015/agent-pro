# Scenario

**Feature**: ExactlyOneMoreSpace classify

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opOccupied
	return nil
}
```
