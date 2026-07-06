# Scenario

**Feature**: RenderSudoersLine formats NOPASSWD sudoers entries

```
Manager.RenderSudoersLine -> "<user> ALL=(root) NOPASSWD: <cmd> [<args>]"
```

## Steps

1. Set `Request.Operation = "render"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "render"
	return nil
}
```