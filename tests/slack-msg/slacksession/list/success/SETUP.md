# Scenario

**Feature**: session list success paths (empty / multi / json / limit)

```
seed sessions.json under HOME
  -> slack-msg session list [...]
  -> sorted human table or --json document
```

## Preconditions

- Home isolated by parent list Setup.
- Fixtures use fixed timestamps for deterministic sort (`sessionListFixtureEntries`).

## Steps

1. Ensure HomeDir isolation still present for unit leaves.
2. Leaf seeds map entries (or empty map) and sets list flags.

## Context

- Sort key: `updated_at` descending.
- `--limit N` applied after sort (data rows only; header excluded).

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.HomeDir == "" {
		return fmt.Errorf("list success leaves require HomeDir isolation from parent Setup")
	}
	return nil
}
```
