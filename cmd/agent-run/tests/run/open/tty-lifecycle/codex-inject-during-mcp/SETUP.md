# Scenario

**Feature**: `--open` injects a Codex prompt while fake TUI still shows Starting MCP servers

```
agent-run run --agent-runner codex-tty --open MCP_INJECT_OK
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  + fake TUI: OpenAI Codex + Starting MCP servers + › , then read stdin
  -> probe STDIN=MCP_INJECT_OK
```

## Preconditions

- Inherits tty-lifecycle INSTANT attach.
- Codex never puts the prompt on argv; inject must land on the PTY.

## Steps

1. Override runner to `codex-tty`.
2. Fake TUI keeps MCP chrome on screen while reading stdin.
3. Assert the prompt was injected (not skipped until MCP ends).

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.OpenInstantAttach = true
	req.Prompt = "MCP_INJECT_OK"
	probe := filepath.Join(req.TempDir, "mcp-inject-probe.txt")
	script := filepath.Join(req.TempDir, "fake-codex-mcp-inject.sh")
	body := fmt.Sprintf(`#!/bin/sh
printf 'OpenAI Codex\n• Starting MCP servers (0/2): slow_30\n› '
# Keep MCP chrome visible while the PTY input endpoint and child stdin settle.
sleep 2
if read -r line; then
  printf 'STDIN=%%s\n' "$line" > %q
else
  printf 'STDIN_COUNT=0\n' > %q
fi
sleep 2
`, probe, probe)
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		return err
	}
	setCodexTTYCommand(req, script)
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Args = []string{"run", "--agent-runner", "codex-tty", "--open", req.Prompt}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
