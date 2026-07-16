# Scenario

**Feature**: headless JSON run with commandcode-tty

```
agent-run run --json --agent-runner commandcode-tty --agent-runner-binary <mock> "Hello"
  -> exit 0, stdout has JSON events with message text, stderr has session id
```

## Steps

1. Run with `--json` and `"Hello"` prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"run", "--json",
		"--agent-runner", "commandcode-tty",
		"--agent-runner-binary", req.MockBinary,
		"Hello",
	}
	return nil
}
```
