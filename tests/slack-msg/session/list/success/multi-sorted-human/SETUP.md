# Scenario

**Feature**: session list human multi-row sorted by updated_at desc

```
two map entries (newer channel + older dm without dir)
  -> session list
  -> header SESSION_ID… + padded columns; newer first; empty dir as -
```

## Steps

1. Seed two entries (older first in file; product re-sorts).
2. Args: session list.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, sessionListFixtureEntries()); err != nil {
		return err
	}
	req.Args = []string{"session", "list"}
	return nil
}
```
