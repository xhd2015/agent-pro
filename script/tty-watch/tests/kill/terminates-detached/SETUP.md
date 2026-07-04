# Scenario

**Feature**: kill terminates a detached session and removes registry entry

```
# detached sleep then kill
harness -> detached sleep -> tty-watch kill -> registry gone, not reachable
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "kill-stop"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```