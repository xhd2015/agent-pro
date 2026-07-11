# Scenario

**Feature**: `--session-id` conflicts with `--session-id-from-prompt` like `--session`

```
agent-run run --session-id X --session-id-from-prompt "p"
  -> error, exit ≠ 0
```

## Steps

1. Parent sets `run --agent-runner fake-codex`; leaf adds both conflicting flags + prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "conflict prompt"
	req.Args = append(req.Args,
		"--session-id", "explicit-id",
		"--session-id-from-prompt",
		req.Prompt,
	)
	return nil
}
```
