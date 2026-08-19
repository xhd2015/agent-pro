# Scenario

**Feature**: active `session_meta.cwd` discovers the rollout transcript before a resume footer

```
fake Codex TUI starts without printing codex resume
  -> rollout JSONL appears with session_meta.cwd = workspace
  -> event_msg.agent_message streams while PTY is still alive
  -> resume footer is printed only after the streamed message window
```

## Preconditions

- The assistant text appears only in the rollout JSONL.
- The fake TUI does not print a resume footer until after the expected stream probe window.

## Steps

1. Schedule `session_meta` creation shortly after `agent-run` starts.
2. Schedule an assistant JSONL line before any resume footer is visible in scrollback.
3. Assert stdout sees the assistant marker before PTY exit.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
	"github.com/xhd2015/doctest/session"
)

const codexActiveCWDText = "JSONL_ACTIVE_CWD_BEFORE_RESUME_FOOTER"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CodexTTYCommand = fakeTUINoResumeUntilLate(5)
	path := ensureCodexTranscriptPath(t, req)
	req.CodexTranscriptSchedules = []CodexTranscriptSchedule{
		{
			Delay: 200 * time.Millisecond,
			Lines: []string{
				codexSessionMetaLine(req.CodexTranscriptSessionID, req.TempDir),
				// Prompt gate: cwd scan needs a matching user message before the
				// late resume footer (fakeTUINoResumeUntilLate) is printed.
				codexUserMessageLine("run ls"),
			},
		},
		{
			Delay: 800 * time.Millisecond,
			Lines: []string{
				codexAgentMessageLine(codexActiveCWDText),
			},
		},
	}
	req.CodexTranscriptPath = path
	req.StreamProbeSubstring = codexActiveCWDText
	req.StreamProbeTimeout = 25 * time.Second
	req.ExecTimeout = 50 * time.Second
	return nil
}
```
