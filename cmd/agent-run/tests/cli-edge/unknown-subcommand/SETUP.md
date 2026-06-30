# Scenario

**Feature**: unknown subcommand exits with code 1

```
agent-run not-a-real-command → exit 1
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Run `agent-run not-a-real-command`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"not-a-real-command"}
	return nil
}
```