# Scenario

**Feature**: draft with one trailing space is exactly-one-more-space

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Before = "hello draft"
	req.After = "hello draft "
	return nil
}
```
