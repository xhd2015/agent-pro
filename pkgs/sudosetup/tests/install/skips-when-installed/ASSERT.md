## Expected

- `EnsureInstalled` succeeds without error.
- `Installed` remains true.
- No `visudo` or `install` runner calls.

## Side Effects

- Sudoers and manifest content unchanged.

## Errors

- None.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Installed, true)
	if hasRunnerCall(resp.RunnerCalls, "sudo", "visudo") {
		t.Fatal("visudo must not run when already installed")
	}
	if hasRunnerCall(resp.RunnerCalls, "sudo", "install") {
		t.Fatal("install must not run when already installed")
	}
}
```