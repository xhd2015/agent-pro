## Expected Output

```
---
version: 3
---
No user prompts found
```

## Expected

- Output contains `No user prompts found` (window-specific wording may add
  detail but must keep this core phrase).
- Trailing newline.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertContains(t, resp.Output, "No user prompts found")
	assertTrailingNewline(t, resp.Output)
}
```
