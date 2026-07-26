## Preconditions
- End-to-end reproduction: grok native lines (thought + tool_started/completed + text)
  are converted to `events.jsonl` via `GrokEventWriter`, then displayed by `traceSession`.
- This is the pipeline gap that causes `doctest agent implement --trace` to omit tool calls.

## Steps
1. Feed mixed grok native JSON lines through `GrokEventWriter` to build events.jsonl content.
2. Create a session dir with that content and call `traceSession`.

```go
import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/grok"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := t.TempDir()
	sessDir := filepath.Join(dir, "sess_test")

	grokLines := []string{
		`{"type":"thought","data":"The user wants me to act as the implementer."}`,
		`{"type":"tool_started","tool_name":"Read"}`,
		`{"type":"tool_completed","tool_name":"Read","duration_ms":1,"outcome":"success"}`,
		`{"type":"tool_started","tool_name":"Grep"}`,
		`{"type":"tool_completed","tool_name":"Grep","duration_ms":2,"outcome":"success"}`,
		`{"type":"text","data":"I'll implement the DOCTEST.md version and layout changes."}`,
	}

	var buf bytes.Buffer
	w := grok.NewGrokEventWriter(&buf)
	for _, line := range grokLines {
		w.WriteGrokLine(line)
	}
	w.Flush()

	req.PreCreateDirs = []string{sessDir}
	req.PreCreateMeta = map[string]string{
		sessDir: `{"explicit_session_id":"test-grok-native-pipeline","agent_runner":"grok","created_at":"2026-06-22T22:36:10Z"}`,
	}
	req.PreCreateEvents = map[string]string{
		sessDir: buf.String(),
	}
	req.SessionID = "test-grok-native-pipeline"
	req.SessionBase = dir
	return nil
}
```