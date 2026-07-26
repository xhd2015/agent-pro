# Scenario

**Feature**: stats errors when session UUID is not found

```
# empty Grok home, no matching summary.json
sessions.Stats(grokHome, unknownUUID) -> error

# error message names the missing session
grok session not found: <id>
```

## Preconditions

- No `summary.json` exists for the requested session id.
- Full UUID is required (no prefix matching).

## Steps

1. Set `req.SessionID` to a UUID that does not exist on disk.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "019f283b-eeee-7eee-eeee-eeeeeeeeeeee"
	return nil
}
```
