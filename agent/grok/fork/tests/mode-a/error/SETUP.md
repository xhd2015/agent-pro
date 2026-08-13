# Scenario

**Feature**: Mode A hard failures do not open a window

```
fork.Main(bare) -> error (no ancestor / no session / empty cwd)
OpenInNewTerminal not called
```

## Preconditions

- Leaves override procs, open files, or session cwd.

## Steps

1. Leaf breaks the happy fixture.
2. Assert returned error + no open.

## Context

- `Main` returns error without `Error:` prefix.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{}
	return nil
}
```
