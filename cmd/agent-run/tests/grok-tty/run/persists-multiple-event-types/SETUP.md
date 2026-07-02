# Scenario

**Feature**: streamed ACP updates persist user + think + tool + assistant AgentEvents

```
temp GROK_HOME session with full ACP sequence in updates.jsonl
  -> events.jsonl under AGENT_RUN_HOME has multiple event types (not one end blob)
```

## Steps

1. Pre-seed `updates.jsonl` with user, thought, tool_call, tool_call_update, assistant chunks.
2. Run with short-hold fake TUI so tailer can read the file before exit.
3. Assert `events.jsonl` contains user message, tool_call, assistant message, and think.

```go
import "testing"

const persistGrokUUID = "33333333-3333-3333-3333-333333333333"

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = persistGrokUUID
	prompt := "persist events"
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, persistGrokUUID, prompt,
		acpAgentThoughtChunk("planning ls output"),
		acpToolCall("call_persist", "execute", "ls"),
		acpToolCallUpdate("call_persist", "completed", "agent\nagents"),
		acpAgentMessageChunk("PERSIST_ASSISTANT_MARKER"),
	)
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIHoldSeconds(2)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, prompt)
	return nil
}
```