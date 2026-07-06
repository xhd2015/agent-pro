## Expected

- `Status.Installed` is true.
- `Status.InstallDetail` references matching drop-in/manifest.

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
	assertEqual(t, "Installed", resp.Status.Installed, true)
	if resp.Status.InstallDetail == "" {
		t.Fatal("InstallDetail should be non-empty when installed")
	}
}
```