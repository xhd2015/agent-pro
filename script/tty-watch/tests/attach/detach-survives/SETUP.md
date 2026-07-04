# Scenario

**Feature**: attach Ctrl-] (`\x1d`) detaches client only; session survives

```
# attach then detach; registry + list still show session
harness -> detached sleep -> tty-watch attach -> \x1d -> exit 0; list includes session
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-detach-survives"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```