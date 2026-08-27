# Scenario

**Feature**: newline-only difference ⇒ Changed true (raw equality)

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Before = "a\nb"
	req.After = "ab"
	return nil
}
```
