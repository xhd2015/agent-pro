## Expected

- At least one file exists under `sessions/<runner>/<sessionID>/`.
- `events.jsonl` is among the files written.
- All paths remain under `AGENT_RUN_HOME`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.FilesWritten) == 0 {
		t.Fatal("expected files after first AppendEvent")
	}
	wantPrefix := filepath.Join(resp.ResolvedHome, "sessions", req.Runner, req.SessionID)
	found := false
	for _, p := range resp.FilesWritten {
		if strings.HasPrefix(p, wantPrefix) {
			found = true
		}
		if strings.HasSuffix(p, "events.jsonl") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected events.jsonl under %q, got files: %v", wantPrefix, resp.FilesWritten)
	}
	AssertHomeOnly(t, resp.ResolvedHome, resp.FilesWritten)
}
```