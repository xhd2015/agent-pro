# Scenario

**Feature**: in-process Handle smoke (help + unknown)

```
# help
Handle(["--help"])
  -> err nil
  -> stdout lists Usage + commands (web, run, sessions, status) + --agent-runner

# unknown
Handle(["not-a-real-command"])
  -> err non-nil (unknown command)
```

## Preconditions

- Calls `agentruncli.Handle` only; no separate `agent-run` binary.
- Stdout/stderr captured around Handle (root Run mutex).
- Leaves set `req.Args`.

## Steps

1. Set harness mode `handle`.
2. Leaf sets Args; Run invokes Handle and fills Stdout/Stderr/ErrString.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = "handle"
	return nil
}
```
