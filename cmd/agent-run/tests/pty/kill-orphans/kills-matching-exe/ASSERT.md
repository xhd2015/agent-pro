## Expected

- Exit code 0.
- Spawned serve PID is no longer alive after kill-orphans.
- Stdout ends with trailing newline `\n` (summary of killed PIDs is OK).

## Side Effects

- Only the test-binary serve is required to die; host serves must be untouched
  (enforced by `--exe` isolation, not asserted via host scan).

## Exit Code

0

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if resp.ServePID <= 0 {
		t.Fatalf("expected spawned serve PID in response")
	}
	// Allow brief settle if kill is asynchronous.
	deadline := time.Now().Add(3 * time.Second)
	alive := resp.ServeAliveAfter
	for alive && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		alive = processAlive(resp.ServePID)
	}
	if alive {
		t.Fatalf("serve pid %d still alive after kill-orphans --exe; stdout:\n%s\nstderr:\n%s",
			resp.ServePID, resp.Stdout, resp.Stderr)
	}
	assertTrailingNewline(t, resp.Stdout, "kill-orphans stdout")
}
```
