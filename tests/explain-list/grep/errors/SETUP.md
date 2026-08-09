# Scenario

**Feature**: invalid --grep / mode-flag combinations hard-fail

```
# empty pattern | both --or and --and | mode without greps
-> stderr "Error: …", non-zero exit
```

## Preconditions

- Leaves set offending Args; sessions optional (errors should fail before list).

## Steps

1. Invoke list with invalid combination.
2. Assert non-zero exit, stderr contains `Error:` and relevant tokens.

## Context

- Product main prints `Error: %v\n` on stderr and exits 1.
- Exact wording flexible; soft substring match on flag/empty semantics.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("grep/errors setup: explain binary not built")
	}
	return nil
}
```
