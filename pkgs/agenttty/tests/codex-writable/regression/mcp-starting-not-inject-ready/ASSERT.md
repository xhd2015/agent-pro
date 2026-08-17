## Expected

- `CheckWritable`: `ready=false`, `state=loading`, reason mentions MCP starting.
- Fixture contains `Starting MCP servers` and a main-chat `›` / `»`.
- `BannerDetected(scrollback, "codex", []string{"CODEX_TTY_BANNER"})` is **false**
  (same predicate `waitForBannerRemote` / `agent-run run` uses before inject).

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureMCPServersStarting))
	if err != nil {
		t.Fatal(err)
	}
	s := string(text)
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "starting mcp servers") {
		t.Fatalf("fixture must contain Starting MCP servers")
	}
	if !strings.Contains(s, "›") && !strings.Contains(s, "\u203a") &&
		!strings.Contains(s, "»") && !strings.Contains(s, "\u00bb") {
		t.Fatalf("fixture must contain main chat › or » (the leak that makes banner true today)")
	}
	assertWritable(t, "mcp-starting-not-inject-ready", resp.Status, false, "loading", "MCP")
	// Desired: run must not treat MCP-boot chrome as inject-ready.
	if agenttty.BannerDetected(text, "codex", []string{"CODEX_TTY_BANNER"}) {
		t.Fatal("BannerDetected must be false while Starting MCP servers (run would inject too early)")
	}
}
```
