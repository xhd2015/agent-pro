## Expected

- `resp.InputBox` is `empty`.
- `CheckWritable` remains `ready=true`, `state=idle`.

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertInputBox(t, resp, err, "empty")
	assertWritableIdle(t, resp, "empty glued snapshot")
}
```
