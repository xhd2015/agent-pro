# Scenario

**Feature**: Ctrl-] (`\x1d`) detaches client; session survives in registry and list

```
# detach client only
harness PTY -> tty-watch sleep -> \x1d -> list still shows session
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-detach"
	req.Detach = true
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```