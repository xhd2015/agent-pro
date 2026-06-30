# Scenario

**Feature**: run requires interactive terminal like attach

```
# non-TTY stdin
piped stdin/stdout -> agent-term run bash -> error: interactive terminal required
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-requires-tty"
	req.StartDaemon = true
	req.RunCommand = []string{"bash"}
	return nil
}
```