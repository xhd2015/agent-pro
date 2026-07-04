# Scenario

**Feature**: tty-watch host stays silent during attach and detach

```
# no session-N printed by host on stdout/stderr during writer attach
harness PTY -> tty-watch sleep -> detach -> no host session id lines
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-silent"
	req.RunCommand = []string{"sleep", "120"}
	return nil
}
```