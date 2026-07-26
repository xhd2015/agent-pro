# Scenario

**Feature**: Handle unknown subcommand returns error

```
# unknown command
Handle(["not-a-real-command"])
  -> error (unknown / unrecognized command)
```

## Preconditions

- Same failure as today's `agent-run not-a-real-command` (error returned from
  dispatch; thin main maps to exit 1 — exit mapping not asserted here).

## Steps

1. Mode `handle`.
2. Args `["not-a-real-command"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = "handle"
	req.Args = []string{"not-a-real-command"}
	return nil
}
```
