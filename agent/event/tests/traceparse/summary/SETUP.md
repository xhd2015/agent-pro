# Scenario

**Feature**: summary helper functions

```
# mode selects which participant runs
doctest harness -> Run Mode=summary -> consolidated trace packages
```

## Preconditions
- Root `Run` handles Mode `summary`.

## Steps
1. Descendant leaves set `req.Mode` to `summary` (and wire-specific fields).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "summary"
	return nil
}
```
