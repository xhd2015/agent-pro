## Expected

- `Escape` succeeds without error.
- Escaped output doubles `"` as `\"`.
- Escaped output doubles `\` as `\\`.
- Result can be embedded in an AppleScript string literal without breaking syntax.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}
	assertEscapedContains(t, resp.Escaped, `\"`)
	assertEscapedContains(t, resp.Escaped, `\\`)
}
```
