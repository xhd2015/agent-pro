# Scenario

**Feature**: empty session store produces a friendly zero-exit message

```
# no sessions under debug config home
explain list -> "No explain sessions yet." (exit 0, trailing newline)
```

## Preconditions

- Config home exists but has no valid session directories (or no sessions/ dir content).

## Steps

1. Leave `req.Sessions` empty.
2. Run `explain list`.
3. Assert friendly empty message, exit 0, trailing newline, no ANSI.

## Context

- Empty store is not an error; users may run `list` before any ask.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("empty setup: explain binary not built (root Setup skipped?)")
	}
	// No sessions seeded — ConfigHome is empty isolation dir from root Setup.
	req.Sessions = nil
	return nil
}
```
