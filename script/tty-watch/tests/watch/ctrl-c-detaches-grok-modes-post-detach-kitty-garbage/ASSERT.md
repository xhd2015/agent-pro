## Expected

- Watch detaches on iTerm kitty Ctrl-C after grok modes and restores the observer TTY.
- Detach cleanup pops grok's kitty keyboard stack (`\x1b[<u`) so iTerm2 stops delivering
  kitty key events for plain typing.
- Post-detach output must not contain kitty protocol fragments (`d0;1:3u`, `a7;1:3u`,
  `0u9;5:3u`, `100;1:3u`, `97;1:3u`, `;1:3u`) or echoed cleanup (`[?0u`).
- Remote session remains reachable; watch exit code 0.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GrokModesSeen {
		t.Fatalf("grok-like terminal modes not observed, output %q", resp.Combined)
	}
	if resp.TimedOut {
		t.Fatal("watch did not detach before post-detach kitty input probe")
	}
	if !resp.TTYCleanupOnDetach {
		t.Fatalf("watch detach missing kitty keyboard pop cleanup (\\x1b[<u]), got %q", resp.Combined)
	}
	post := resp.PostDetachOutput
	if post == "" {
		t.Fatal("no post-detach output captured after watch detach")
	}
	if ttywatchtest.PostDetachOutputHasKittyGarbage(post) {
		t.Fatalf("post-detach shell echoed kitty protocol garbage after grok watch detach, got %q", post)
	}
	idx := strings.Index(resp.Combined, "WATCH_ENDED")
	if idx >= 0 {
		after := resp.Combined[idx:]
		if strings.Contains(after, "[?0u") {
			t.Fatalf("post-detach output echoed kitty disable CSI smear [?0u, got %q", after)
		}
	}
}
```