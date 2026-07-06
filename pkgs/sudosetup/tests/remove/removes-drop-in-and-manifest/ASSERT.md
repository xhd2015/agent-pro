## Expected

- `Remove` succeeds.
- Runner records `sudo rm`, manifest deletion, and `sudo -k`.
- `Installed` is false afterward.

## Side Effects

- Drop-in and manifest removed from FS.

## Errors

- None.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Installed, false)
	if !hasRunnerCall(resp.RunnerCalls, "sudo", "rm") {
		t.Fatal("expected sudo rm call")
	}
	if !hasRunnerCall(resp.RunnerCalls, "sudo", "-k") {
		t.Fatal("expected sudo -k call")
	}
	if resp.SudoersContent != "" {
		t.Fatal("sudoers drop-in should be removed")
	}
	if resp.ManifestJSON != nil && len(resp.ManifestJSON) > 0 {
		t.Fatal("manifest should be removed")
	}
}
```