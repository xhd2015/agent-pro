# Scenario

**Feature**: combining `--session` with `--auto-session-id` fails

```
agent-run run --agent-runner fake-codex --session my-id --auto-session-id "x"
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
		"--auto-session-id",
		req.Prompt,
	)
	return nil
}
```
