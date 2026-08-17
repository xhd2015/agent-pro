---
label: e2e,codex
explanation: >-
  Live Codex TUI with llm-mock-run-codex + MCP config must not be inject-ready
  while Starting MCP servers is on screen (agent-run run would inject too early).
---

## Expected

- Isolated `config.toml` contains the hang `[mcp_servers.slowinit_*]` blocks.
- Snapshot contains `Starting MCP servers` (real TUI, not a checked-in fixture).
- `CheckWritable` is `loading` (MCP starting).
- `BannerDetected` is **false** so `run` waits (RED today: `codex` + `›` leaks true).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if !strings.Contains(resp.ConfigTOML, "[mcp_servers.slowinit_01]") {
		t.Fatalf("CODEX_HOME/config.toml missing hang MCP servers (harness):\n%s\ndetach stderr:\n%s",
			resp.ConfigTOML, resp.DetachStderr)
	}
	if !resp.SawMCPBoot {
		t.Fatalf("did not observe live Starting MCP servers within poll\nlast snapshot:\n%s\ndetach stderr:\n%s\nconfig:\n%s",
			resp.Snapshot, resp.DetachStderr, resp.ConfigTOML)
	}
	if !strings.Contains(strings.ToLower(resp.Snapshot), "starting mcp") {
		t.Fatalf("snapshot missing Starting MCP servers:\n%s", resp.Snapshot)
	}
	if resp.Writable.Ready || resp.Writable.State != "loading" {
		t.Fatalf("CheckWritable want loading/not-ready; got ready=%v state=%q reason=%q\nsnapshot:\n%s",
			resp.Writable.Ready, resp.Writable.State, resp.Writable.Reason, resp.Snapshot)
	}
	if !strings.Contains(resp.Writable.Reason, "MCP") {
		t.Fatalf("CheckWritable reason %q want MCP starting\nsnapshot:\n%s",
			resp.Writable.Reason, resp.Snapshot)
	}
	if resp.BannerReady {
		t.Fatal("BannerDetected must be false while Starting MCP servers (run would inject too early)")
	}
}
```
