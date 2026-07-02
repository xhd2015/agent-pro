# Scenario

**Feature**: discovered grok UUID is persisted as `runner_session_id` in agent-run meta.json

```
temp GROK_HOME session with known UUID
  -> after run, meta.json runner_session_id equals grok session UUID
```

## Steps

1. Seed fake grok session dir with fixed UUID.
2. Set `AGENT_RUN_GROK_TTY_GROK_SESSION_ID` hook (optional fast path).
3. Run with quick-respond fake TUI; read `meta.json` after completion.

```go
import "testing"

const storeGrokUUID = "550e8400-e29b-41d4-a716-446655440000"

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = storeGrokUUID
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, storeGrokUUID, "store id",
		acpAgentMessageChunk("STORE_ID_ASSISTANT"),
	)
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIHoldSeconds(1)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, "store id")
	return nil
}
```