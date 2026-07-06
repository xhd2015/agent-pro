## Expected

- `Status.Installed` is false.
- `Status.InstallDetail` mentions command mismatch.

## Side Effects

- None.

## Errors

- None from `Run`.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Status.Installed, false)
	assertContains(t, resp.Status.InstallDetail, "command")
}
```