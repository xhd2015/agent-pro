# Scenario

**Feature**: send without required args fails without stdout id

```
agent-run send -> exit 1, stderr usage error, no msg_N stdout
```

## Steps

1. Set `req.Action = "missing-args"`.
2. Set `req.SendArgs = []string{"send"}`.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "missing-args"
	req.SendArgs = []string{"send"}
	return nil
}
```