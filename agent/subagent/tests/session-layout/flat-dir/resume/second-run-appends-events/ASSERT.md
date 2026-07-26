## Expected

- First `subagent.Run` (doctest `Run`) succeeds.
- Second `invokeRun` with `SecondMockConfigPath` succeeds.
- `events.jsonl` line count after second run is greater than after first run.
- Host meta fields (`id`, `task_title`, `project_name`, `agent_session_id`) unchanged.
- `opencode_session_id` is set to `inner_resume_run1` after runs.

## Side Effects

- Events append in place; meta.json is merged not replaced.

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

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("first Run failed: %v", err)
	}

	eventsPath := filepath.Join(req.SessionDir, "events.jsonl")
	afterFirst := countJSONLLines(eventsPath)
	if afterFirst == 0 {
		t.Fatalf("expected events after first run, got 0 lines")
	}

	req.MockConfigPath = req.SecondMockConfigPath
	_, err2 := invokeRun(t, req)
	if err2 != nil {
		t.Fatalf("second Run failed: %v", err2)
	}

	afterSecond := countJSONLLines(eventsPath)
	if afterSecond <= afterFirst {
		t.Fatalf("events not appended: first=%d second=%d", afterFirst, afterSecond)
	}

	metaPath := filepath.Join(req.SessionDir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parse meta: %v", err)
	}

	checks := map[string]string{
		"id":               "20260625-fixture-resume",
		"task_title":       "resume title",
		"project_name":     "fixture-resume",
		"agent_session_id": "gen_layout_resume_test",
		"opencode_session_id": "inner_resume_run1",
	}
	for key, want := range checks {
		got, _ := meta[key].(string)
		if got != want {
			t.Fatalf("meta.%s = %q, want %q", key, got, want)
		}
	}
}```
