# Scenario

**Feature**: short `echo` shows exactly two visible lines without scrollback smear

```
fresh terminal -> tty-watch run echo yes -> visible lines:
  yes
  [Terminal exited]
```

Late writer attach replays `\x1b[?1049l\x1b[0m` scrollback prefix, which on a
cleared terminal paints many blank lines before `yes`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-echo-clean-output"
	req.RunCommand = []string{"echo", "yes"}
	return nil
}
```