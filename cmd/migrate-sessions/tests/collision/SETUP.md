# Scenario

**Feature**: migration collision when same session id exists under two runners

```
sessions/codex/same + sessions/grok/same -> keep newer bare same; rename loser same__codex
```

## Preconditions

- Q1: keep newer `updated_at` (fallback `created_at`) at bare `{id}`; rename others to `{id}__{runner}`.

## Steps

1. Leaf seeds colliding nested sessions and runs migrator.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SeedMode = "collision"
	return nil
}
```
