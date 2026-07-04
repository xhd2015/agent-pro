## Expected

- Watch output shows grok-like terminal modes (`1049h`, `?u`) before Ctrl-C.
- iTerm-style kitty Ctrl-C (`\x1b[99;5u`) detaches watch within 3s.
- Before exiting, watch restores the observer terminal:
  - leave alt-screen (`\x1b[?1049l`)
  - pop grok kitty keyboard protocol push (`\x1b[<u`)
  - disable mouse tracking (`\x1b[?1000l`, `\x1b[?1002l`, `\x1b[?1003l`, `\x1b[?1006l`)
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
		t.Fatal("watch did not detach on iTerm kitty Ctrl-C within 3s after grok mode preamble")
	}
	if !resp.TTYCleanupOnDetach {
		t.Fatalf("watch detach left grok terminal modes active on observer TTY (missing cleanup sequences), got %q", resp.Combined)
	}
	if !resp.RegistryExists {
		t.Fatalf("registry missing after watch detach, session %s", resp.SessionID)
	}
	if !resp.SessionRunning {
		t.Fatalf("remote session %s not reachable after watch Ctrl-C detach", resp.SessionID)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("watch exit code = %d after Ctrl-C detach, want 0", resp.ExitCode)
	}
}
```