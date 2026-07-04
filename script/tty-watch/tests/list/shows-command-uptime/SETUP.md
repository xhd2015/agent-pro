# Scenario

**Feature**: list prints session id, joined command argv, and human uptime

```
# detached sleep session
harness -> detached sleep -> tty-watch list -> id + sleep + uptime
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "list-fields"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```