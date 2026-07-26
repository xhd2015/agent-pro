# Scenario

**Feature**: Cursor merge lifecycle

```
# mode selects which participant runs
doctest harness -> Run Mode=parse_messages -> consolidated trace packages
```

## Preconditions
- Root `Run` handles Mode `parse_messages`.

## Steps
1. Descendant leaves set `req.Mode` to `parse_messages` (and wire-specific fields).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "parse_messages"
	return nil
}
```
