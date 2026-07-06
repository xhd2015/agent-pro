## Expected

- `EnsureInstalled` returns an error mentioning visudo.
- `Installed` remains false.
- Manifest file is not written.

## Side Effects

- No manifest JSON on disk.

## Errors

- Non-nil error from `Run`.

## Exit Code

- Failure.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertError(t, err)
	assertEqual(t, "Installed", resp.Installed, false)
	if resp.ManifestJSON != nil && len(resp.ManifestJSON) > 0 {
		t.Fatal("manifest must not be written when visudo fails")
	}
	assertContains(t, err.Error(), "visudo")
}
```