## Expected

- Error containing `not found` and the session id.
- `session_bookmarks.json` is absent (no silent create).

## Errors

- `not found` + session id

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "not found", req.SessionID)
	assertPathMissing(t, storePath(req.AgentProHome))
}
```
