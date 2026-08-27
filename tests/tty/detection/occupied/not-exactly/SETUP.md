# Scenario

**Feature**: placeholder collapse is not exactly-one-more-space

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Before = "Ask Codex to do anything"
	req.After = ""
	return nil
}
```
