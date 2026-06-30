# Scenario

**Feature**: PTY spawner runs shell or arbitrary commands

```
# spawn path
session manager -> pty.Start(cmd) -> child process in PTY
```

## Steps

1. Set `req.Phase` for spawn scenarios.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "spawn-shell"
	return nil
}
```