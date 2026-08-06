# Scenario

**Feature**: GetBookmark on missing entry errors

```
GetBookmark(runner=grok, unknown-id)
  -> error not found
```

## Preconditions

- No matching store entry.
- Op=show.

## Steps

1. Set unknown SessionID; Runner=grok; Op=show.

```go
import "testing"

const missingShowSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee66"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = missingShowSessionID
	req.Runner = "grok"
	req.Op = "show"
	return nil
}
```
