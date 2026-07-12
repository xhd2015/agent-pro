# Scenario

**Feature**: `-e FOO` without `=` is a validation error

```
run --agent-runner grok-tty -e FOO "hi" -> exit ≠ 0; clear error
```

## Steps

1. Run with `-e FOO` (no equals sign) and a prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"run",
		"--agent-runner", "grok-tty",
		"-e", "FOO",
		"hi",
	}
	return nil
}
```
