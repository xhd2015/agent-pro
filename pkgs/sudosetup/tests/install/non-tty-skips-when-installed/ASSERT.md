# Scenario

## Expected

- `EnsureInstalled` succeeds with no error.
- No `visudo` or `install` runner calls.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if hasRunnerCall(resp.RunnerCalls, "sudo", "visudo") {
		t.Fatal("visudo must not run when already installed")
	}
	if hasRunnerCall(resp.RunnerCalls, "sudo", "install") {
		t.Fatal("install must not run when already installed")
	}
}
```