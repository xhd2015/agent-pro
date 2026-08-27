# Scenario

**Feature**: empty before + one space is not occupied

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Before = ""
	req.After = " "
	return nil
}
```
