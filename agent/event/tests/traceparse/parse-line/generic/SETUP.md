# Scenario

**Feature**: generic assistant fallback adapter

```
# mode selects which participant runs
doctest harness -> Run Mode=parse_line -> consolidated trace packages
```

## Preconditions
- Root `Run` handles Mode `parse_line`.

## Steps
1. Descendant leaves set `req.Mode` to `parse_line` (and wire-specific fields).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "parse_line"
	return nil
}
```
