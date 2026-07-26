# Scenario

**Feature**: list with zero sessions prints empty-store message

```
# isolated empty config home
explain list -> stdout: "No explain sessions yet.\n" ; exit 0
```

## Preconditions

- No `sessions/*` dirs under debug config home.

## Steps

1. Keep `req.Sessions` empty.
2. Args remain `["list"]`.

## Context

- Exact wording locked to `No explain sessions yet.` per requirement example.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list"}
	req.Sessions = nil
	return nil
}
```
