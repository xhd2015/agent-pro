# Scenario

**Bug**: after Skip, residual banner still contains Update available and was treated as blocking

```
03b-menu-dismissed.snapshot.txt -> IsBlocking=false; writable reason ≠ codex update available
```

## Preconditions

Signed fixture `03b-menu-dismissed.snapshot.txt`. Still has `model: loading`.

## Steps

1. `FixtureFile=03b-menu-dismissed.snapshot.txt`.

## Context

PROTOCOL `confirm_skip` + writable narrow gate.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FixtureFile = "03b-menu-dismissed.snapshot.txt"
	return nil
}
```
