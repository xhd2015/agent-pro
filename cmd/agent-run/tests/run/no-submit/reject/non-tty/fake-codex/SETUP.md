# Scenario

**Feature**: `--open --no-submit` with `fake-codex` fails as non-TTY

```
agent-run run --agent-runner fake-codex --open --no-submit "x"
  -> exit ≠ 0
  -> stderr explains --open / TTY requirement (not unrecognized flag only)
```

## Steps

1. Invoke run with `--open`, `--no-submit`, `fake-codex`, and a short prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "fake-codex"
	req.Prompt = "x"
	req.Args = []string{
		"run",
		"--agent-runner", "fake-codex",
		"--open",
		"--no-submit",
		req.Prompt,
	}
	return nil
}
```
