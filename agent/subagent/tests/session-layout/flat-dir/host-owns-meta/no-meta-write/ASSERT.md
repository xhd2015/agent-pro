## Expected

- `subagent.Run` succeeds.
- `meta.json` bytes are **identical** to pre-run snapshot (subagent did not write).
- `events.jsonl` exists and is non-empty.
- `OnAgentComplete` was called (host would persist separately).

## Side Effects

- Events written under flat dir; meta untouched by subagent.

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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	metaPath := filepath.Join(req.SessionDir, "meta.json")
	after, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta.json: %v", err)
	}
	if !bytes.Equal(req.MetaBytesBeforeRun, after) {
		t.Fatalf("meta.json changed under HostOwnsMeta\nbefore:\n%s\nafter:\n%s", req.MetaBytesBeforeRun, after)
	}

	eventsPath := filepath.Join(req.SessionDir, "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil || len(data) == 0 {
		t.Fatalf("events.jsonl missing or empty: %v", err)
	}

	if !req.CallbackCalled {
		t.Fatal("OnAgentComplete was not called")
	}
}```
