---
label: e2e,codex
---

## Expected

- Exit code 0.
- Probe file contains `STDIN=MCP_INJECT_OK` (injected while MCP chrome was up).
- `bind.json` exists under the session dir (in_progress resolved to ok or failed).

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)
	probe, err := os.ReadFile(filepath.Join(req.TempDir, "mcp-inject-probe.txt"))
	if err != nil {
		t.Fatalf("read inject probe: %v\nstderr:\n%s", err, resp.Stderr)
	}
	got := string(probe)
	if !strings.Contains(got, "STDIN=MCP_INJECT_OK") {
		t.Fatalf("--open must inject while Starting MCP servers is on screen; probe:\n%s\nstderr:\n%s", got, resp.Stderr)
	}
	matches, _ := filepath.Glob(filepath.Join(req.Home, "sessions", "*", "bind.json"))
	if len(matches) == 0 {
		t.Fatalf("expected bind.json under %s/sessions/*\nstderr:\n%s", req.Home, resp.Stderr)
	}
}
```
