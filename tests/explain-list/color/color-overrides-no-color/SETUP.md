# Scenario

**Feature**: --color overrides NO_COLOR

```
# NO_COLOR=1 + explain list --color -> ANSI still on
```

## Preconditions

- One short session.
- Both `--color` and `NO_COLOR=1`.

## Steps

1. Seed fixture.
2. Args `list --color`; EnvExtra `NO_COLOR=1`.
3. Assert ANSI present (bold cyan / green / dim).

## Context

- Spec: `--color` wins over everything including `NO_COLOR`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list", "--color"}
	req.EnvExtra = []string{"NO_COLOR=1"}
	req.Sessions = []SessionSeed{colorFixtureSession()}
	return nil
}
```
