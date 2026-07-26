# Scenario

**Feature**: `--session-id` is an alias for `--session` on `agent-run run` (AR1)

```
# AR1
agent-run run --agent-runner fake-codex --session-id my-task "hi"
  -> sessions/fake-codex/my-task/
  -> meta.session_id == my-task
  # same as --session my-task
```

## Steps

1. Run with `--session-id my-task` (not `--session`) and prompt `hi`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "fake-codex"
	req.Prompt = "hi"
	req.Args = append(req.Args,
		"--agent-runner", "fake-codex",
		"--session-id", "my-task",
		req.Prompt,
	)
	return nil
}
```
