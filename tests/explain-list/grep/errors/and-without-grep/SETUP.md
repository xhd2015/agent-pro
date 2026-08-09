# Scenario

**Feature**: --and without any --grep is a hard error

```
# explain list --and
-> non-zero exit; stderr Error: … --grep
```

## Preconditions

- No greps; only `--and`.

## Steps

1. Args: `list --and`.
2. Assert hard error mentioning grep.

## Context

- Symmetric to `--or` without greps.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--and"}
	req.Sessions = nil
	return nil
}
```
