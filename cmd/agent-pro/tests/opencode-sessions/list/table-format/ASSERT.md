## Expected

- `FormatListTable` output contains column headers `SESSION ID`, `LAST ACTIVE`, `TITLE`, `MSGS`, `CWD`.
- Relative times appear for each session: `just now`, `5m ago`, `2h ago`.
- Output includes session ids, titles, message counts, and cwd paths.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "SESSION ID")
	assertContains(t, resp.Output, "LAST ACTIVE")
	assertContains(t, resp.Output, "TITLE")
	assertContains(t, resp.Output, "MSGS")
	assertContains(t, resp.Output, "CWD")

	assertContains(t, resp.Output, "ses_table_alpha")
	assertContains(t, resp.Output, "ses_table_beta")
	assertContains(t, resp.Output, "ses_table_gamma")

	assertContains(t, resp.Output, "just now")
	assertContains(t, resp.Output, "5m ago")
	assertContains(t, resp.Output, "2h ago")

	assertContains(t, resp.Output, "Alpha refactor")
	assertContains(t, resp.Output, "Beta bugfix")
	assertContains(t, resp.Output, "Gamma cleanup")

	assertContains(t, resp.Output, "/tmp/project-a")
	assertContains(t, resp.Output, "/tmp/project-b")
	assertContains(t, resp.Output, "/tmp/project-c")
}
```