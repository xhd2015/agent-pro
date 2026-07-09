# Scenario

**Feature**: explicit `--session` sets non-TTY storage session id

```
agent-run run --agent-runner fake-codex --session my-task "hi"
  -> sessions/fake-codex/my-task/
  -> meta.session_id == my-task
```

## Steps

1. Run with `--session my-task` and prompt `hi`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "fake-codex"
	req.Prompt = "hi"
	req.Args = append(req.Args,
		"--agent-runner", "fake-codex",
		"--session", "my-task",
		req.Prompt,
	)
	return nil
}
```
