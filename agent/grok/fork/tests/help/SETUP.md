# Scenario

**Feature**: grok-fork help lists two-mode flags and omits new-terminal

```
fork.Main(--help|-h) -> usage on stdout -> err=nil
```

## Preconditions

- Help is Mode-independent (no ancestor walk required).

## Steps

1. Leaf sets `-h` or `--help`.
2. Assert required flag names and absence of `-n` / `--new-terminal`.

## Context

- Exit 0 from library (`Main` returns nil).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	return nil
}
```
