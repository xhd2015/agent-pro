# Scenario

**Feature**: serve logs HTTP session create requests to stderr

```
# request observability
REST POST /api/terminal/sessions -> daemon stderr logs method + path
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "serve-logs-on-create"
	req.StartDaemon = true
	return nil
}
```