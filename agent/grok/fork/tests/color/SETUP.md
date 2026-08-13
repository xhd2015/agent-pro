# Scenario

**Feature**: `--color` / `--no-color` / auto-on-pipe gate ANSI on Mode A output

```
seed Mode A fixture
fork.Main([--color|--no-color|…]) -> green Opened / gray labels / no ANSI
```

## Preconditions

- Same ancestor fixture as Mode A.
- Writers are buffers (pipe) so auto = off.

## Steps

1. Seed Mode A session/chain.
2. Leaf adds color flags.

## Context

- Green `Opened` (`\x1b[32m`); gray labels (`\x1b[90m`).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedModeA(t, req)
	return nil
}
```
