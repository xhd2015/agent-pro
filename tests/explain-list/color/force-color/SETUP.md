# Scenario

**Feature**: --color forces ANSI even on non-TTY

```
# explain list --color (piped stdout)
-> Q bold cyan, A bold green, header dim
```

## Preconditions

- One short session; no NO_COLOR.

## Steps

1. Seed color fixture.
2. Args: `list --color`.
3. Assert ANSI sequences on labels/meta.

## Context

- `--color` wins over non-TTY auto-off.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list", "--color"}
	req.EnvExtra = nil
	req.Sessions = []SessionSeed{colorFixtureSession()}
	return nil
}
```
