# Scenario

**Feature**: combining `--session` with `--session-id-from-prompt` fails

```
agent-run run --agent-runner fake-codex --session my-id --session-id-from-prompt "x"
  -> exit ≠ 0
  -> stderr explains mutual exclusion / conflict
```

## Steps

1. Invoke run with both flags and a short prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "x"
	req.Args = append(req.Args,
		"--session", "my-id",
		"--session-id-from-prompt",
		req.Prompt,
	)
	return nil
}
```
