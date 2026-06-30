# Scenario

**Feature**: unknown `--agent-runner` exits with code 1 and stderr hint

```
agent-run run --agent-runner totally-bogus-runner "hi" → exit 1, stderr mentions unknown
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Run `agent-run run --agent-runner totally-bogus-runner "hi"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"run",
		"--agent-runner", "totally-bogus-runner",
		"hi",
	}
	return nil
}
```