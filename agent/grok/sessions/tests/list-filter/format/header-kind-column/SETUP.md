# Scenario

**Feature**: FormatListTable header places KIND immediately after SESSION ID

```
one main session
  -> ListWithOptions + FormatListTable
  -> first line has SESSION ID then KIND then LAST ACTIVE
```

## Preconditions

- WantFormat=true.
- At least one session so header is printed.

## Steps

1. Write one plain main session.
2. WantFormat=true, Limit=10.
3. Assert header column order.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.WantFormat = true
	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-10*time.Minute), cwdA, "header main", listSessionOpts{})
	return nil
}
```
