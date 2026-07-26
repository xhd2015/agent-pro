# Scenario

**Feature**: legacy ai-critic WS create-on-connect with name and cwd

```
# no session_id query
WS ?name=&cwd= -> create shell -> session_id JSON message
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "ws-create-on-connect"
	req.Name = "compat-shell"
	req.Cwd = absTempDir(t)
	return nil
}
```