# Scenario

**Feature**: Handle top-level --help lists subcommands and flags

```
# help smoke
Handle(["--help"])
  -> stdout includes Usage, web, run, sessions, status, --agent-runner
  -> err == nil
```

## Preconditions

- Same help surface as today's `cmd/agent-run` top-level help (zero behavior change).

## Steps

1. Mode `handle`.
2. Args `["--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = "handle"
	req.Args = []string{"--help"}
	return nil
}
```
