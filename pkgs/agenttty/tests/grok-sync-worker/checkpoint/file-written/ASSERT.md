## Expected

- `grok-sync.json` exists under session dir.
- `updates_offset` > 0 and equals current `updates.jsonl` file size (fully processed).
- `grok_session_id` and `updates_path` populated in checkpoint.

## Side Effects

- Checkpoint written after events appended (offset reflects processed bytes).

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureGrokSync: %v", resp.EnsureErr)
	}
	path := grokSyncJSONPath(req.SessionDir)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("grok-sync.json missing: %v", err)
	}
	cp := readGrokSyncCheckpoint(t, req.SessionDir)
	if cp.UpdatesOffset <= 0 {
		t.Fatalf("updates_offset must be > 0, got %d", cp.UpdatesOffset)
	}
	fileSize := updatesFileSize(t, req.UpdatesPath)
	if cp.UpdatesOffset != fileSize {
		t.Fatalf("updates_offset %d != updates.jsonl size %d", cp.UpdatesOffset, fileSize)
	}
	if cp.GrokSessionID != req.GrokSessionID {
		t.Fatalf("grok_session_id: got %q want %q", cp.GrokSessionID, req.GrokSessionID)
	}
	if cp.UpdatesPath != req.UpdatesPath {
		t.Fatalf("updates_path: got %q want %q", cp.UpdatesPath, req.UpdatesPath)
	}
	if !resp.CheckpointOK {
		t.Fatal("Response.CheckpointOK false")
	}
	if resp.Checkpoint.UpdatesOffset > 0 && resp.Checkpoint.UpdatesOffset != cp.UpdatesOffset {
		t.Fatalf("response checkpoint offset %d != file %d",
			resp.Checkpoint.UpdatesOffset, cp.UpdatesOffset)
	}
}
```
