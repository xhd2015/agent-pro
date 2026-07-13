# Scenario

**Feature**: without --color, non-TTY auto leaves color off

```
# explain list (piped stdout, no NO_COLOR, no --color) -> plain
```

## Preconditions

- One short session.
- EnvExtra empty (NO_COLOR stripped by harness).

## Steps

1. Seed fixture; plain `list`.
2. Assert no ANSI.

## Context

- exec.Command captures pipes → not a TTY → auto color off.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list"}
	req.EnvExtra = nil
	req.Sessions = []SessionSeed{colorFixtureSession()}
	return nil
}
```
