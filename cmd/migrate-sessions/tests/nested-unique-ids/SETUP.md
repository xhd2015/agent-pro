# Scenario

**Feature**: migrate nested sessions with unique ids across runners

```
sessions/fake-codex/a + sessions/fake-opencode/b -> sessions/a + sessions/b
```

## Preconditions

- Nested layout only; no id collisions.
- Events and meta must be preserved under flat dirs.

## Steps

1. Grouping marks seed mode nested unique.
2. Leaf seeds two nested sessions and runs migrator.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SeedMode = "nested_unique"
	return nil
}
```
