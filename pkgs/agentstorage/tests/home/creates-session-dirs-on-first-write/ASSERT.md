## Expected

- At least one file exists under `sessions/<sessionID>/`.
- `events.jsonl` is among the files written.
- No file path contains `sessions/<runner>/<sessionID>` nested layout.
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
	wantPrefix := filepath.Join(resp.ResolvedHome, "sessions", req.SessionID)
	nestedPrefix := filepath.Join(resp.ResolvedHome, "sessions", req.Runner, req.SessionID)
	foundEvents := false
	for _, p := range resp.FilesWritten {
		if strings.HasPrefix(p, nestedPrefix) {
			t.Fatalf("unexpected nested runner path %q (want flat under %q)", p, wantPrefix)
		}
		if strings.HasPrefix(p, wantPrefix) && strings.HasSuffix(p, "events.jsonl") {
			foundEvents = true
		}
	}
	if !foundEvents {
		t.Fatalf("expected events.jsonl under %q, got files: %v", wantPrefix, resp.FilesWritten)
	}
	AssertHomeOnly(t, resp.ResolvedHome, resp.FilesWritten)
}
```