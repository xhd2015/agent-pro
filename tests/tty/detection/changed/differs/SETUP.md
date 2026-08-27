# Scenario

**Feature**: different content ⇒ Changed true

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Before = "version one"
	req.After = "version two"
	return nil
}
```
