## Expected

- `resp.InputBox` is `unknown`.
- Empty bytes are not treated as an empty composer (that requires a glyph).

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertInputBox(t, resp, err, "unknown")
}
```
