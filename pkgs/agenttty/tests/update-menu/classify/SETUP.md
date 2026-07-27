# Scenario

**Feature**: classify signed update-modal snapshots (menu vs banner)

```
fixture text -> IsBlockingUpdateMenu + UpdateMenuSelection + CheckWritable
```

## Preconditions

Signed fixtures under `testdata/update-modal-skip/`. No live Codex.

## Steps

1. Leaf sets `FixtureFile` and optional `StripModelLoading`.

## Context

Fast CI leaves (no labels). Drive production classifier surface.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	return nil
}
```
