## Expected

- No API error.
- Returned JSON is what Launch wrote.
- Launch ran once (production wait, not a Wait hook).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "JSON", resp.JSON, `{"status":"ready"}`)
	assertEqual(t, "LaunchCalls", resp.LaunchCalls, 1)
}
```
