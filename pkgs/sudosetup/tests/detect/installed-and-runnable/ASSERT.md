## Expected

- `Status.Installed` is true.
- `Status.CanRunNonInteractive` is true.
- `Status.Verdict` indicates permanent NOPASSWD is configured.

## Side Effects

- Runner probes `sudo -n` with command path.

## Errors

- None from `Run`.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Status.Installed, true)
	assertEqual(t, "CanRunNonInteractive", resp.Status.CanRunNonInteractive, true)
	if !detailContains(resp.Status.Verdict, "permanent") && !detailContains(resp.Status.Verdict, "configured") {
		t.Fatalf("Verdict should indicate permanent install, got %q", resp.Status.Verdict)
	}
}
```