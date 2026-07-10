# Scenario

**Feature**: session-id-from-prompt with non-TTY runner (fake-codex) sets storage id only

```
agent-run run --agent-runner fake-codex --session-id-from-prompt "prompt"
  -> sessions/fake-codex/<slug-ts…>/
  -> no terminal registry required
```

## Preconditions

- `fake-codex` on PATH (root Setup).

## Steps

1. Prefix `--agent-runner fake-codex`.
2. Leaves add prompts and optional collision seeds.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "fake-codex"
	req.Args = append(req.Args, "--agent-runner", "fake-codex")
	return nil
}
```
