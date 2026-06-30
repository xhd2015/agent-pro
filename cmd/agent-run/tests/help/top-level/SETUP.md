# Scenario

**Feature**: top-level `--help` lists subcommands and flags

```
agent-run --help → stdout lists web, run, sessions, status, --agent-runner
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Run `agent-run --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```