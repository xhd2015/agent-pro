## Preconditions
- **Reproduces**: `doctest agent design --trace` shows per-word thinking blocks for grok sessions.
- Grok streams `thought` events word-by-word; each becomes an `ActionThink` line in events.jsonl.
- `traceSession` prints each think event as a separate numbered `💭` block.

## Steps
1. Set up a session dir with events.jsonl containing per-word `ActionThink` events.
2. Deltas simulate grok streaming: "The", " user", " wants", " me", " to", " act".
3. Call `traceSession`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := t.TempDir()
	sessDir := filepath.Join(dir, "sess_test")

	eventsContent := `{"type":"think","text":"The"}
{"type":"think","text":" user"}
{"type":"think","text":" wants"}
{"type":"think","text":" me"}
{"type":"think","text":" to"}
{"type":"think","text":" act"}
`

	req.PreCreateDirs = []string{sessDir}
	req.PreCreateMeta = map[string]string{
		sessDir: `{"explicit_session_id":"test-grok-thought-streaming","created_at":"2026-06-22T16:40:30Z"}`,
	}
	req.PreCreateEvents = map[string]string{
		sessDir: eventsContent,
	}
	req.SessionID = "test-grok-thought-streaming"
	req.SessionBase = dir
	return nil
}
```