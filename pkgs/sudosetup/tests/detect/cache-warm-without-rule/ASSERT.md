## Expected

- `Status.Installed` is false.
- `Status.CacheWarm` is true.
- `Status.Verdict` indicates cache-only (rule not installed).

## Side Effects

- Runner receives `sudo -n true` probe.

## Errors

- None from `Run`.

## Exit Code

- Success.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Status.Installed, false)
	assertEqual(t, "CacheWarm", resp.Status.CacheWarm, true)
	if !detailContains(resp.Status.Verdict, "cache") {
		t.Fatalf("Verdict should mention cache-only, got %q", resp.Status.Verdict)
	}
	if !hasRunnerCall(resp.RunnerCalls, "sudo", "-n", "true") {
		t.Fatal("expected sudo -n true probe")
	}
}
```