## Expected

- `subagent.Run` succeeds without replacing entire meta.json.
- `meta.id` remains `20260625-fixture-a-merged`.
- `meta.task_title` remains `keep this title`.
- `meta.project_name` remains `fixture-a`.
- `meta.opencode_session_id` is set to `inner_merged_sess`.
- `meta.agent_session_id` remains `gen_layout_merged_test`.

## Side Effects

- `events.jsonl` created; no `messages.jsonl` (MessagesPath empty).

## Errors

- None.

## Exit Code

N/A

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	metaPath := filepath.Join(req.SessionDir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta.json: %v", err)
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parse meta.json: %v\n%s", err, string(data))
	}

	checks := map[string]string{
		"id":                 "20260625-fixture-a-merged",
		"task_title":         "keep this title",
		"project_name":       "fixture-a",
		"agent_session_id":   "gen_layout_merged_test",
		"opencode_session_id": "inner_merged_sess",
	}
	for key, want := range checks {
		got, _ := meta[key].(string)
		if got != want {
			t.Fatalf("meta.%s = %q, want %q\nfull meta: %v", key, got, want, meta)
		}
	}

	if _, err := os.Stat(filepath.Join(req.SessionDir, "events.jsonl")); err != nil {
		t.Fatalf("events.jsonl missing: %v", err)
	}
	if fileExists(filepath.Join(req.SessionDir, "messages.jsonl")) {
		t.Fatalf("unexpected messages.jsonl with MessagesPath empty")
	}
}
```
