# Scenario

**Feature**: --or without any --grep is a hard error

```
# explain list --or
-> non-zero exit; stderr Error: … --grep
```

## Preconditions

- No greps; only `--or`.

## Steps

1. Args: `list --or`.
2. Assert hard error mentioning grep.

## Context

- Mode flags only make sense with at least one pattern.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--or"}
	req.Sessions = nil
	return nil
}
```
