## Expected

- `RenderSudoersLine` returns:
  `testuser ALL=(root) NOPASSWD: /tmp/cache/remote-agent-sudo-poc/hello.sh`
- No trailing arg pattern beyond command path.

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
	want := "testuser ALL=(root) NOPASSWD: " + req.Rule.Command
	assertEqual(t, "RenderedLine", resp.RenderedLine, want)
}
```