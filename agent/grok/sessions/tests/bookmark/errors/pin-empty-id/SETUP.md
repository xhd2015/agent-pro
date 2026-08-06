# Scenario

**Feature**: pin with empty session id errors

```
BookmarkGrok("", ...)
  -> error (not found or empty id); no store write
```

## Preconditions

- SessionID is empty string.
- Op=pin.

## Steps

1. Set SessionID=""; Op=pin.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = ""
	req.Op = "pin"
	req.NilOpts = true
	return nil
}
```
