# Scenario

**Feature**: `--open` starts keep-alive PTY, snapshot returns response text

```
agent-run run --open --agent-runner commandcode-tty --agent-runner-binary <mock> "Hello"
  -> exit 0, stderr commandcode-tty: session-N, snapshot contains "Hello"
```

## Steps

1. Run with `--open` and `"Hello"` prompt.
2. Wait for response to appear.
3. Take snapshot.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	withCommandcodeOpenTestEnv(req)
	req.Args = []string{
		"run",
		"--open",
		"--agent-runner", "commandcode-tty",
		"--agent-runner-binary", req.MockBinary,
		"Hello",
	}
	return nil
}
```
