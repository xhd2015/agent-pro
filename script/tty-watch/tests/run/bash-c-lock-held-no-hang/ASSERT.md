## Expected

When another process holds the registry flock, `tty-watch run bash -c 'echo yes'`
must **not hang forever**.

Correct contract:
- Return within the short harness budget (~3s) without `TimedOut`.
- Exit non-zero with an error that mentions the registry lock (or "busy" / "in use"),
  **or** succeed if the implementation no longer needs an exclusive flock for this path.
- Must not block indefinitely on `flock(LOCK_EX)`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.TimedOut {
		t.Fatalf("tty-watch run bash -c 'echo yes' hung while registry .lock held (timed out after %s); expected prompt lock error, output %q",
			resp.Elapsed, resp.Combined)
	}
	// Success is acceptable only if the run no longer requires exclusive flock.
	if resp.ExitCode == 0 {
		combined := resp.Combined
		if combined == "" {
			combined = resp.Stdout
		}
		if !strings.Contains(combined, "yes") {
			t.Fatalf("exit 0 without yes output while lock held: %q", combined)
		}
		return
	}
	// Non-zero: must explain lock contention rather than a generic hang/timeout.
	combined := resp.Combined + "\n" + resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "lock") &&
		!strings.Contains(lower, "busy") &&
		!strings.Contains(lower, "in use") &&
		!strings.Contains(lower, "timeout") {
		t.Fatalf("expected lock-related error when registry flock held, exit %d output %q",
			resp.ExitCode, combined)
	}
}
```
