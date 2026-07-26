# Scenario

**Feature**: print.FormatTraceLine integration

```
# mode selects which participant runs
doctest harness -> Run Mode=print -> consolidated trace packages
```

## Preconditions
- Root `Run` handles Mode `print`.

## Steps
1. Descendant leaves set `req.Mode` to `print` (and wire-specific fields).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "print"
	return nil
}
```
