## Expected

- First `Run` succeeds.
- Second `invokeRun` with `ResumeInnerSessionID=inner_host_resume_run1` succeeds.
- `events.jsonl` line count increases after second run.
- `meta.json` bytes unchanged from pre-first-run snapshot.
- Second run `OnAgentComplete` invoked (host would update meta externally).

## Side Effects

- Events append; meta never written by subagent.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"bytes"
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
		t.Fatalf("expected events after first run")
	}

	req.MockConfigPath = req.SecondMockConfigPath
	req.ResumeInnerSessionID = "inner_host_resume_run1"
	req.CallbackCalled = false
	_, err2 := invokeRun(t, req)
	if err2 != nil {
		t.Fatalf("second Run failed: %v", err2)
	}

	afterSecond := countJSONLLines(eventsPath)
	if afterSecond <= afterFirst {
		t.Fatalf("events not appended: first=%d second=%d", afterFirst, afterSecond)
	}

	metaPath := filepath.Join(req.SessionDir, "meta.json")
	after, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	if !bytes.Equal(req.MetaBytesBeforeRun, after) {
		t.Fatalf("meta.json changed across resume runs under HostOwnsMeta\nbefore:\n%s\nafter:\n%s", req.MetaBytesBeforeRun, after)
	}
	if !req.CallbackCalled {
		t.Fatal("OnAgentComplete not called on second run")
	}
}```
