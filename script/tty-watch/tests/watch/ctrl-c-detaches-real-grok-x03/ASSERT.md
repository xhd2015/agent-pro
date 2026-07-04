---
label: real-grok, slow
---

## Expected

- Real grok alt-screen modes observed before Ctrl-C.
- Injected `\x03` detaches watch within 3s; grok session survives.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GrokModesSeen {
		t.Fatalf("grok terminal modes not observed, output %q", resp.Combined)
	}
	if !strings.Contains(resp.Combined, "\x1b[?1049h") {
		t.Fatalf("watch output missing alt-screen enable, got %q", resp.Combined)
	}
	if resp.TimedOut {
		t.Fatal("watch did not detach on \\x03 within 3s after real grok alt-screen")
	}
	if !resp.RegistryExists || !resp.SessionRunning {
		t.Fatalf("grok session %s not alive after detach", resp.SessionID)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("watch exit code = %d, want 0", resp.ExitCode)
	}
}
```