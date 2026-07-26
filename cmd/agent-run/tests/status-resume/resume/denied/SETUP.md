# Scenario

**Feature**: resume denied paths (exit ≠ 0)

```
not-exited (live → hint send, not already-in-use)
  | unbound | missing-session | missing-prompt
  | --no-submit without --open
  -> exit ≠ 0 + clear error
```

## Steps

1. Leaf seeds the denying state and runs resume.
2. Assert non-zero exit and error wording.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Grouping marker for deny-path leaves (exit 1 expected).
	req.Mode = ""
	return nil
}
```
