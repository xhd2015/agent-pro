# Scenario

**Feature**: --limit above 100 is capped at 100

```
# 3 sessions; explain list --limit 200 -> title limit 100, 3 shown
```

## Preconditions

- Few sessions (3) so only the effective limit value is under test, not disk bulk.

## Steps

1. Seed 3 sessions.
2. Args: `list --limit 200`.
3. Assert title uses limit 100 (not 200); all 3 shown.

## Context

- Cap applies to the limit parameter; shown count is min(total, capped limit).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--limit", "200"}
	req.Sessions = seedNSessions(3, 12)
	return nil
}
```
