# Scenario

**Feature**: list-specific help documents flags

```
# explain list --help / -h -> usage mentions --limit and --color
```

## Preconditions

- Implementer registers list help text including `--limit` and `--color`.

## Steps

1. Leaves run `list --help` or equivalent.
2. Assert exit 0 and flag names in output (stdout or stderr — whichever flags lib uses).

## Context

- less-gen `Help` typically prints and exits 0 in a subprocess.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("help setup: explain binary not built")
	}
	return nil
}
```
