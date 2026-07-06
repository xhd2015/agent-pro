## Expected

- `Status.Installed` is false.
- `Status.InstallDetail` mentions drop-in not present.
- `Status.CacheWarm` is false (default runner).
- `Status.Verdict` indicates password is required.

## Side Effects

- No runner calls beyond detect probes.

## Errors

- None from `Run`.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Status.Installed, false)
	assertContains(t, resp.Status.InstallDetail, "drop-in")
	assertEqual(t, "CacheWarm", resp.Status.CacheWarm, false)
	assertContains(t, resp.Status.Verdict, "password")
}
```