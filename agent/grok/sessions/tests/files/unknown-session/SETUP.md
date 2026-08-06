# Scenario

**Feature**: ListSessionFiles errors for unknown session id

```
empty sessions tree
-> ListSessionFiles(unknown-id)
-> error grok session not found
```

## Preconditions

- No session directory for the requested id.

## Steps

1. Set SessionID to a missing UUID.
2. Do not write fixtures.

```go
import "testing"

const unknownFilesSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee88"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = unknownFilesSessionID
	return nil
}
```
