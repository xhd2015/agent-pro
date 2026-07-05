---
label: slow
explanation: waits up to ~10s for headless ctrl-c grace window
---

## Expected

- After SIGINT, stderr contains a **single** line `waiting for program to exit...` appearing after ~1s.
- Headless parent exits 1 after force-kill (~10s total).
- Registry session removed after shutdown.

## Exit Code

1

```go
import (
	"strings"
	"testing"
	"time"
)

const headlessWaitingLine = "waiting for program to exit..."

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("expected exit 1 after force-kill, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	count := strings.Count(resp.Stderr, headlessWaitingLine)
	if count != 1 {
		t.Fatalf("expected exactly one waiting line on stderr, count=%d stderr=%q", count, resp.Stderr)
	}
	if resp.AltStderr == "" {
		t.Fatal("harness should record waiting line latency in AltStderr")
	}
	waitingAfter, parseErr := time.ParseDuration(resp.AltStderr)
	if parseErr != nil {
		t.Fatalf("parse waiting latency %q: %v", resp.AltStderr, parseErr)
	}
	if waitingAfter < 800*time.Millisecond {
		t.Fatalf("waiting line appeared too early: %s", waitingAfter)
	}
	if waitingAfter > 4*time.Second {
		t.Fatalf("waiting line appeared too late: %s", waitingAfter)
	}
	if resp.Elapsed < 9*time.Second {
		t.Fatalf("expected ~10s grace window, elapsed=%s", resp.Elapsed)
	}
	if resp.RegistryExists || len(resp.RegistryIDs) != 0 {
		t.Fatalf("registry should be gone after force-kill, ids=%v", resp.RegistryIDs)
	}
}
```