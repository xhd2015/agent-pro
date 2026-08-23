# Scenario

**Feature**: agent enrich path (NoEnrich=false)

```
NoEnrich=false -> per-session Idle class + ResolveFromPID -> optional Agents entry
```

## Steps

1. Ensure enrich is enabled for the subtree.
2. Child branches split on Idle class and resolve outcome.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.NoEnrich = false
	return nil
}
```
