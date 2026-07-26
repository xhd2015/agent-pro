# Scenario

**Feature**: --limit N restricts number of sessions shown

```
# 5 sessions; explain list --limit 3 -> 3 newest only
```

## Preconditions

- Five valid sessions.

## Steps

1. Seed 5 sessions.
2. Args: `list --limit 3`.
3. Assert 3 shown / limit 3; newest three present; oldest two absent.

## Context

- Explicit positive limit overrides default 10.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--limit", "3"}
	req.Sessions = seedNSessions(5, 8)
	return nil
}
```
