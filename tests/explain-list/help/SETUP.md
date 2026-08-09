# Scenario

**Feature**: list-specific help documents flags

```
# explain list --help / -h -> usage mentions --limit, --grep, --or, --and, --color
```

## Preconditions

- Implementer registers list help text including `--limit`, `--grep`, `--or`,
  `--and`, and `--color`.

## Steps

1. Leaves run `list --help` or equivalent.
2. Assert exit 0 and flag names in output (stdout or stderr — whichever flags lib uses).

## Context

- less-gen `Help` typically prints and exits 0 in a subprocess.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("help setup: explain binary not built")
	}
	return nil
}
```
