## Expected

- `EnsureInstalled` succeeds.
- Manifest JSON has `username=testuser`, `command=/opt/homebrew/bin/sing-box`,
  `args_pattern=run -c *`.
- `Installed` is true after install.

## Side Effects

- Manifest file created at `ManifestPath`.

## Errors

- None.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Installed, true)
	if resp.ManifestJSON == nil {
		t.Fatal("manifest JSON missing")
	}
	assertEqual(t, "username", resp.ManifestJSON["username"], "testuser")
	assertEqual(t, "command", resp.ManifestJSON["command"], "/opt/homebrew/bin/sing-box")
	assertEqual(t, "args_pattern", resp.ManifestJSON["args_pattern"], "run -c *")
}
```