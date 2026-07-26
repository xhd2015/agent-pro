# Scenario

**Feature**: migrator no-ops when home is already flat

```
sessions/.layout v2 + sessions/<id>/ -> migrate-sessions -> exit 0; no destructive rewrite
```

## Preconditions

- Flat sessions already present with `.layout` version 2.

## Steps

1. Leaf seeds flat layout and runs migrator.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SeedMode = "already_flat"
	return nil
}
```
