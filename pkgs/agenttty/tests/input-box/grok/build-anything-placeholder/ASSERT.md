## Expected

- `resp.InputBox` is `empty`.
- Placeholder `Build anything` after `❯` is chrome, not user draft text.

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertInputBox(t, resp, err, "empty")
}
```
