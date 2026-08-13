# Scenario

**Feature**: with workspace `AGENTS.md` (Codex first user message), `--open`
still binds `runner_session_id`; auto-send-or-resume keeps the same Codex id.

Crime scene: `~/seatalk-local-bot/AGENTS.md` is stored as user[1]; SeaTalk
inject is user[2]. Discover prompt-match against user[1] fails → unbound →
ModeRun second uuid.

```
workspace AGENTS.md + SeaTalk-shaped inject
  -> llm-mock-run-codex --open
  -> meta.runner_session_id bound (same uuid as the one rollout)
  -> /exit
  -> --auto-send-or-resume --open FOLLOW_UP
  -> still one Codex uuid
```

Expect **RED** until Discover matches the inject among the first few user
messages (not only the first).

## Steps

1. Inherit root Setup (isolated homes, binaries, skip without `codex`).
2. Write `AGENTS.md` into `req.Workspace`; set SeaTalk-shaped `OpenPrompt`.
3. Mock exchanges match `SeaTalk local-bot session open` and `FOLLOW_UP`.
4. Shared Run: open → end → auto-send-or-resume.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = "codex-open-agents-md-s1"
	req.OpenPrompt = "SeaTalk local-bot session open\n" +
		"session-id: codex-open-agents-md-s1\n" +
		"playbook: ~/.spl/seatalk-local-bot/sessions/codex-open-agents-md-s1/SYSTEM.md\n" +
		"cli: spl seatalk local-bot get --session-id codex-open-agents-md-s1\n" +
		"cli: spl seatalk local-bot reply --session-id codex-open-agents-md-s1\n" +
		"Chat history (1 message):\n" +
		"message_id=m-probe : @Marcuus debug-resume-probe-1924 ping\n"
	req.FollowupPrompt = "FOLLOW_UP"
	if req.Workspace == "" {
		t.Fatal("workspace empty (root Setup should set it)")
	}
	agents := filepath.Join(req.Workspace, "AGENTS.md")
	body := "# Mad-max agent brief (local-bot)\n\n" +
		"**Route work via skills** — do not invent runbooks.\n"
	if err := os.WriteFile(agents, []byte(body), 0644); err != nil {
		return err
	}
	// Substring match: first user request includes AGENTS.md and/or the inject.
	const mock = `{
  "exchanges": [
    {
      "request": {"role": "user", "content": "SeaTalk local-bot session open", "index": -1},
      "response": {"content": "OPEN_MOCK_REPLY", "finish_reason": "stop"}
    },
    {
      "request": {"role": "user", "content": "FOLLOW_UP", "index": -1},
      "response": {"content": "FOLLOW_UP_MOCK_REPLY", "finish_reason": "stop"}
    },
    {
      "request": {"role": "user", "content": "", "index": -1},
      "response": {"content": "ACK_FALLBACK", "finish_reason": "stop"}
    }
  ]
}
`
	if req.MockConfigFile == "" {
		t.Fatal("MockConfigFile empty")
	}
	return os.WriteFile(req.MockConfigFile, []byte(mock), 0644)
}
```
