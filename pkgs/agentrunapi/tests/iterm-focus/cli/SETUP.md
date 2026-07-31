# Scenario

**Feature**: agent-run focus CLI via agentruncli.RunFocus

```
RunFocus(args, stdout, stderr) -> help | focus path with trailing newline policy
Handle must dispatch "focus" (not unknown command)
```

## Preconditions

- `agentruncli.RunFocus(args, stdout, stderr io.Writer) error` exists (L2 writers).
- `Handle` switch includes `"focus"` (implementer wires to `RunFocus`).

## Steps

1. Leaves set `Phase=cli` and `CLIArgs`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "cli"
	return nil
}
```
