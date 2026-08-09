# Scenario

**Feature**: --or and --and together is a hard error

```
# explain list --grep x --or --and
-> non-zero exit; stderr Error: …
```

## Preconditions

- At least one `--grep` so the conflict is mode-vs-mode, not mode-without-grep.

## Steps

1. Args: `list --grep x --or --and`.
2. Assert hard error.

## Context

- Mutually exclusive combine flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "x", "--or", "--and"}
	req.Sessions = nil
	return nil
}
```
