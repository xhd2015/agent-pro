# Scenario

**Feature**: `less-flags` StopOnFirstArg — command argv tokens are never parsed as flags

```
# --not-a-flag is passed to sh -c literally, not tty-watch
harness -> tty-watch run --detach sh -c 'echo --not-a-flag' -> watch shows token
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-detach-stop-on-first-arg"
	req.RunCommand = []string{"sh", "-c", "echo --not-a-flag"}
	return nil
}
```