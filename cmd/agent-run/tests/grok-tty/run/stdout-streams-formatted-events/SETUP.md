# Scenario

**Feature**: stdout streams formatted user, tool, and assistant events from synthetic updates.jsonl

```
temp GROK_HOME session with full ACP sequence
  -> stdout shows 💬 user line, ⚡ RUN tool line, 💬 assistant line
```

## Steps

1. Pre-seed `updates.jsonl` with user, tool_call, tool_call_update, assistant chunks.
2. Run with short-hold fake TUI so tailer emits formatted stdout before exit.
3. Assert stdout contains formatted event markers (not silent until scrollback fallback).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	formattedGrokUUID      = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	formattedUserText      = "formatted stream events"
	formattedAssistantText = "FORMATTED_ASSISTANT_OUT"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = formattedGrokUUID
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, formattedGrokUUID, formattedUserText,
		acpToolCall("call_formatted", "execute", "ls"),
		acpToolCallUpdate("call_formatted", "completed", "agent\nagents"),
		acpAgentMessageChunk(formattedAssistantText),
	)
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIHoldSeconds(2)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, formattedUserText)
	return nil
}
```