# Scenario

**Feature**: Codex rollout JSONL assistant records stream before PTY exit

```
fake Codex TUI prints resume UUID then sleeps
  -> rollout JSONL receives assistant line while PTY still running
  -> agent-run stdout receives marker before fake TUI exits
```

## Preconditions

- The fake TUI remains alive long enough to distinguish streaming from end-of-run scrollback fallback.

## Steps

1. Seed only `session_meta` before starting.
2. Schedule an `event_msg.agent_message` append after the resume line can be discovered.
3. Assert stdout sees the marker while the PTY process is still sleeping.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
	"github.com/xhd2015/doctest/session"
)

const codexPreExitStreamText = "JSONL_STREAM_BEFORE_PTY_EXIT"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CodexTTYCommand = fakeTUICodexResumeThenSleep(5)
	seedCodexTranscript(t, req)
	req.CodexTranscriptSchedules = []CodexTranscriptSchedule{
		{
			Delay: 800 * time.Millisecond,
			Lines: []string{
				codexAgentMessageLine(codexPreExitStreamText),
			},
		},
	}
	req.StreamProbeSubstring = codexPreExitStreamText
	req.StreamProbeTimeout = 12 * time.Second
	req.ExecTimeout = 35 * time.Second
	return nil
}
```
