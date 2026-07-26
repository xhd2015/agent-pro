# Scenario

**Bug**: empty GROK_HOME + chrome must not end discovery at ~1s via scrollback false completion

```
empty GROK_HOME + real-like chrome + --keep-tty
  -> persistentTurnComplete from chrome must be ignored for grok-tty
  -> discovery polls >3s before bind error surfaces
```

Mirrors `script/debug/grok-tty-discovery-cancel` default `chrome-cancel` scenario.

## Steps

1. Use empty `GROK_HOME` temp dir (no session dirs).
2. Configure real-like chrome hook holding 30s via `LLM_MOCK_RUN_GROK_COMMAND`.
3. Run `agent-run run --keep-tty`; poll `events.jsonl` until resolve error appears.
4. Assert think→error gap exceeds 3s (not ~1.2s context canceled).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "chrome-wait-exceeds-discovery-window"
	req.SessionID = "grok_discovery_chrome_wait"
	req.Prompt = chromeWaitPrompt
	req.GrokHome = filepath.Join(req.TempDir, "empty-grok-home")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	stripEnvPrefix(req, "GROK_HOME=")
	req.Env = append(req.Env, "GROK_HOME="+req.GrokHome)

	configureLLMMockChromeEnv(t, req, req.Prompt, req.ChromeHoldSeconds)
	req.ExecTimeout = 110 * time.Second
	return nil
}
```