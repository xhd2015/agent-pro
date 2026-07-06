## Expected

- `RenderSudoersLine` returns:
  `testuser ALL=(root) NOPASSWD: /opt/homebrew/bin/sing-box run -c *`

## Side Effects

- None (pure render).

## Errors

- None.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	want := "testuser ALL=(root) NOPASSWD: /opt/homebrew/bin/sing-box run -c *"
	assertEqual(t, "RenderedLine", resp.RenderedLine, want)
}
```