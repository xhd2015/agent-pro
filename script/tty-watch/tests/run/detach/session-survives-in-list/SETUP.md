# Scenario

**Feature**: registry session stays live after detach parent exits

```
# detach parent exits; serve child keeps registry entry
harness -> tty-watch run --detach sleep 120 -> exit 0 -> list shows sleep 120
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-detach-session-survives-in-list"
	req.RunCommand = []string{"sleep", "120"}
	return nil
}
```