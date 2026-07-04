## Expected

- Watch output shows grok-like terminal modes (`1049h`, `?u`) before Ctrl-C.
- Kitty-encoded Ctrl-C (`\x1b[3;5u`) detaches watch within 3s.
- Remote session remains registered and TCP-reachable.
- Exit code 0 on clean detach.

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
		t.Fatalf("grok-like terminal modes not observed, output %q", resp.Combined)
	}
	if !strings.Contains(resp.Combined, "\x1b[?1049h") {
		t.Fatalf("watch output missing alt-screen enable, got %q", resp.Combined)
	}
	if !strings.Contains(resp.Combined, "\x1b[?u") {
		t.Fatalf("watch output missing kitty keyboard protocol enable, got %q", resp.Combined)
	}
	if resp.TimedOut {
		t.Fatal("watch did not detach on kitty Ctrl-C within 3s after grok mode preamble")
	}
	if !resp.RegistryExists {
		t.Fatalf("registry missing after watch detach, session %s", resp.SessionID)
	}
	if !resp.SessionRunning {
		t.Fatalf("remote session %s not reachable after watch kitty Ctrl-C detach", resp.SessionID)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("watch exit code = %d after kitty Ctrl-C detach, want 0", resp.ExitCode)
	}
}
```