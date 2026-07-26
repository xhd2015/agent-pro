## Expected

- `FormatListTable` output contains column header `MSGS`.
- Output includes `42` for the session with `num_chat_messages=42`.
- Existing columns `SESSION ID`, `LAST ACTIVE`, `TITLE`, and `CWD` remain present.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "MSGS")
	assertContains(t, resp.Output, "SESSION ID")
	assertContains(t, resp.Output, "LAST ACTIVE")
	assertContains(t, resp.Output, "TITLE")
	assertContains(t, resp.Output, "CWD")
	assertContains(t, resp.Output, "01900005-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	assertContains(t, resp.Output, "42")
	assertContains(t, resp.Output, "Session with messages")
}
```