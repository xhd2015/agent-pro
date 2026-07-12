# Scenario

**Feature**: dry-run prints migration plan without moving files

```
nested home + migrate-sessions --dry-run -> plan on stdout; nested tree unchanged
```

## Preconditions

- Nested layout exists.
- `--dry-run` must not write flat dirs or `.layout` as final layout (no destructive moves).

## Steps

1. Leaf seeds nested sessions and runs with `--dry-run`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	return nil
}
```
