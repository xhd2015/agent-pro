# Scenario

**Feature**: invalid UTF-8 open PROMPT — real `llm-mock-run-grok` + real `grok`

Two leaves (3s budgets; no soft-bind oracle):

1. **direct-llm-mock-crash** — run `llm-mock-run-grok` alone with bad argv; must
   crash (real grok `env.rs`) within 3s.
2. **agent-run-normalized** — same payload via `agent-run` → `llm-mock-run-grok`;
   must **not** crash for 3s; snapshot shows the message.

## Preconditions

- Real `grok` on `PATH`.
- Built sibling `llm-mock` next to `llm-mock-run-grok`.
- No `AGENT_RUN_GROK_TTY_COMMAND` / `LLM_MOCK_RUN_GROK_COMMAND`.
- No probe binary / harness overwrite of `llm-mock-run-grok`.

```go
import (
	"runtime"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/doctest/session"
)

const (
	invalidUTF8Budget       = 3 * time.Second
	invalidUTF8AgentSession = "utf8-agent-run-ok"
)

// incidentOpenPrompt: truncated multi-byte seq (0xE5 0x92) — seatalk incident class.
func incidentOpenPrompt() string {
	// Raw incomplete UTF-8 mid body (latin1 round-trip keeps bytes for Go string).
	raw := append([]byte(nil), []byte(
		"SeaTalk local-bot session open\n"+
			"session-id: seatalk-local-bot-test\n"+
			"First message from master:\n"+
			"在checkout场景 SPL PC rule的逻辑")...)
	raw = append(raw, 0xe5, 0x92)
	raw = append(raw, []byte("edgment (1s)\nalice@example.com: check why the bot is broken\n")...)
	return string(raw)
}

func buildLLMMockSibling(t *testing.T, req *Request) error {
	t.Helper()
	if err := buildLLMMockRunGrok(t, req); err != nil {
		return err
	}
	llmMock := filepath.Join(filepath.Dir(req.LLMMockRunGrok), "llm-mock")
	if _, err := os.Stat(llmMock); err == nil {
		return nil
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", llmMock, "./agent/llm/llm-mock")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock: %w\n%s", err, string(out))
	}
	return nil
}

func requireRealGrok(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skipf("real grok not on PATH: %v", err)
	}
}

func stripGrokHooks(req *Request) {
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	req.Env = withoutEnvKey(req.Env, "LLM_MOCK_RUN_GROK_COMMAND")
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID")
}

func containsGrokEnvPanic(s string) bool {
	return strings.Contains(s, "panicked at library/std/src/env.rs") ||
		strings.Contains(s, "called `Result::unwrap()`")
}

func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "deadline") || strings.Contains(msg, "timeout")
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requireRealGrok(t)
	if err := buildLLMMockSibling(t, req); err != nil {
		return err
	}
	req.Prompt = incidentOpenPrompt()
	if utf8.ValidString(req.Prompt) {
		return fmt.Errorf("fixture: expected invalid UTF-8 open prompt")
	}
	stripGrokHooks(req)
	req.ExecTimeout = invalidUTF8Budget
	return nil
}
```
