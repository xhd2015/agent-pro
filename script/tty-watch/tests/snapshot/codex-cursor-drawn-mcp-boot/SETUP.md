# Scenario

**Bug**: `tty-watch snapshot codex` on real codex truncates top warning text, smears
incremental MCP boot redraws, and leaks kitty protocol CSI fragments.

```
# real codex pattern: ?2026h + CUP (no 2J), warning before alt-screen, MCP status churn
harness -> run --session-id codex (codex-cursor-drawn sh) -> detach -> snapshot codex
```

Reproduces user report where `tty-watch run --session-id codex codex` shows the full
deprecation warning, tip, box UI, and MCP errors, but `tty-watch snapshot codex`
starts mid-warning (`//developers.openai.com/...`), stacks MCP boot lines, and leaks
`[<row;col;23M` fragments.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "snapshot-codex-cursor-drawn-mcp-boot"
	return nil
}
```