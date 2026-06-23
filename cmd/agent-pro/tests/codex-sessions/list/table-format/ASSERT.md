## Expected

- `FormatListTable` output contains column headers `SESSION ID`, `STARTED`, `CWD`.
- Output includes the session UUID and CWD `/tmp/project-a`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "SESSION ID")
	assertContains(t, resp.Output, "STARTED")
	assertContains(t, resp.Output, "CWD")
	assertContains(t, resp.Output, "01900004-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	assertContains(t, resp.Output, "/tmp/project-a")
}
```