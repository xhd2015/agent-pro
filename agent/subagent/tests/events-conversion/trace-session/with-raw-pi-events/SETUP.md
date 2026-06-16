## Preconditions
- **After fix**: events.jsonl contains raw pi events which are properly converted via ConvertRawLine (Bug A) and FromPi (Bug B).
- Each `message_update` event carries the full accumulated text in `message.content[0].text` and a small delta in `assistantMessageEvent.delta`.
- After the fix, the trace formatter outputs only the delta (not the accumulated text), preventing duplication.

## Steps
1. Set up a session dir with events.jsonl containing raw pi message_update events.
2. Each event has growing accumulated text but small deltas.
3. Call `traceSession`.
4. Verify "Hello" appears only once (no duplication).

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    dir := t.TempDir()
    sessDir := filepath.Join(dir, "sess_test")

    // Raw pi events - each message_update has the full accumulated text
    eventsContent := `{"type":"message_start","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}
{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"Hello world"}]},"assistantMessageEvent":{"type":"text_delta","delta":" world"}}
{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"Hello world from pi"}]},"assistantMessageEvent":{"type":"text_delta","delta":" from pi"}}
`

    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"test-raw-pi","created_at":"2026-06-16T10:00:00Z"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: eventsContent,
    }
    req.SessionID = "test-raw-pi"
    req.SessionBase = dir
    return nil
}
```
