# Scenario

**Feature**: empty --grep pattern is a hard error

```
# explain list --grep ""
-> non-zero exit; stderr Error: … non-empty / empty --grep
```

## Preconditions

- No sessions required.

## Steps

1. Args: `list --grep` with empty string value.
2. Assert error path.

## Context

- Empty pattern must not silently match everything.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", ""}
	req.Sessions = nil
	return nil
}
```
