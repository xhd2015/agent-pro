## Preconditions
- **Reproduces**: `doctest agent implement --session-id <id> --trace` shows "fractional message logs".
- pi emits per-token streaming deltas via `message_update` events. Each delta is a single word or
  punctuation character. When traced, each delta becomes its own "ASSISTANT" formatted block,
  resulting in extremely fragmented output.

## Steps
1. Set up a session dir with events.jsonl containing pi `message_update` events with per-token deltas.
2. Delatas simulate real pi streaming: "Let", " me", " start", " by", " understanding".
3. Call `traceSession`.
4. Verify the output is fractional: each tiny delta gets its own "ASSISTANT" header.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    dir := t.TempDir()
    sessDir := filepath.Join(dir, "sess_test")

    // Simulate pi streaming: each delta is a single word/token.
    // This reproduces the "fractional message logs" bug where
    // each tiny delta produces a separate ASSISTANT block.
    eventsContent := `{"type":"message_start","message":{"role":"assistant","content":[{"type":"text","text":"Let"}]}}
{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"Let me"}]},"assistantMessageEvent":{"type":"text_delta","delta":" me"}}
{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"Let me start"}]},"assistantMessageEvent":{"type":"text_delta","delta":" start"}}
{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"Let me start by"}]},"assistantMessageEvent":{"type":"text_delta","delta":" by"}}
{"type":"message_update","message":{"role":"assistant","content":[{"type":"text","text":"Let me start by understanding"}]},"assistantMessageEvent":{"type":"text_delta","delta":" understanding"}}
{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"Let me start by understanding"}]}}
`

    req.PreCreateDirs = []string{sessDir}
    req.PreCreateMeta = map[string]string{
        sessDir: `{"explicit_session_id":"test-streaming","created_at":"2026-06-16T10:00:00Z"}`,
    }
    req.PreCreateEvents = map[string]string{
        sessDir: eventsContent,
    }
    req.SessionID = "test-streaming"
    req.SessionBase = dir
    return nil
}
```
