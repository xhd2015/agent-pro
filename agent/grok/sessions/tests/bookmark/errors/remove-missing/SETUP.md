# Scenario

**Feature**: RemoveBookmark on missing entry errors

```
# empty or unrelated store
RemoveBookmark(grok, unknown-id)
  -> error not found
```

## Preconditions

- Store missing or does not contain the id.
- Op=remove.

## Steps

1. Set unknown SessionID; Runner=grok; Op=remove.

```go
import "testing"

const missingRemoveSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee77"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = missingRemoveSessionID
	req.Runner = "grok"
	req.Op = "remove"
	return nil
}
```
