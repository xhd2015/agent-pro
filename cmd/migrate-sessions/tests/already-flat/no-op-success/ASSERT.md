## Expected

- Exit code 0.
- `sessions/sess_x` still present with original events content.
- `.layout` still version 2.
- No collision renames invented.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitZero(t, resp)
	assertDirExists(t, filepath.Join(req.Home, "sessions", "sess_x"))
	data, err := os.ReadFile(filepath.Join(req.Home, "sessions", "sess_x", "events.jsonl"))
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if !strings.Contains(string(data), "keep-me") {
		t.Fatalf("events rewritten unexpectedly: %s", data)
	}
	if v := layoutVersion(t, req.Home); v != 2 {
		t.Fatalf(".layout version = %d want 2", v)
	}
	// no accidental rename leftovers
	ents, _ := os.ReadDir(filepath.Join(req.Home, "sessions"))
	for _, e := range ents {
		if strings.Contains(e.Name(), "__") {
			t.Fatalf("unexpected renamed session on already-flat home: %s", e.Name())
		}
	}
}
```
