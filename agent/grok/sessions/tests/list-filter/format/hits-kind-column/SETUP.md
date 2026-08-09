# Scenario

**Feature**: FormatListTableWithHits also places KIND after SESSION ID

```
one main session
  -> ListWithOptions + FormatListTableWithHits (empty Hits, color never)
  -> header: SESSION ID  KIND  LAST ACTIVE …
```

## Preconditions

- WantFormatHits=true (Run wraps sessions as SessionMatch).
- Hits path must stay in parity with FormatListTable column layout.

## Steps

1. Write one plain main session.
2. WantFormatHits=true.
3. Assert KIND column order on hits table header.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 10
	req.WantFormatHits = true
	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-10*time.Minute), cwdA, "hits main", listSessionOpts{})
	return nil
}
```
