# Scenario

**Feature**: `-e =bar` (empty key) is a validation error

```
run --agent-runner grok-tty -e =bar "hi" -> exit ≠ 0; clear error
```

## Steps

1. Run with `-e =bar` (empty key before `=`) and a prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"run",
		"--agent-runner", "grok-tty",
		"-e", "=bar",
		"hi",
	}
	return nil
}
```
