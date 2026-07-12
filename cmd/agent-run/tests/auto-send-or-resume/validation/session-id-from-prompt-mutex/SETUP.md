# Scenario

**Feature**: `--auto-send-or-resume` is mutually exclusive with `--session-id-from-prompt`

```
agent-run run --auto-send-or-resume --session-id-from-prompt "hello"
  -> exit 1; mutual exclusion error
```

## Steps

1. Pass both flags with a prompt (no explicit --session-id).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id-from-prompt",
		"hello mutex",
	}
	return nil
}
```
