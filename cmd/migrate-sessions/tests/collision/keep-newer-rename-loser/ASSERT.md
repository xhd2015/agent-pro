## Expected

- Exit code 0.
- Bare `sessions/shared` is the newer (fake-opencode) session; events contain `from-opencode`.
- Loser at `sessions/shared__fake-codex` with events `from-codex` and meta.runner `fake-codex`.
- Report/log mentions rename (stdout or stderr).
- `.layout` version 2.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitZero(t, resp)

	bare := filepath.Join(req.Home, "sessions", "shared")
	loser := filepath.Join(req.Home, "sessions", "shared__fake-codex")
	assertDirExists(t, bare)
	assertDirExists(t, loser)

	if got := readMetaRunner(t, filepath.Join(bare, "meta.json")); got != "fake-opencode" {
		t.Fatalf("bare shared runner = %q want fake-opencode (newer)", got)
	}
	if got := readMetaRunner(t, filepath.Join(loser, "meta.json")); got != "fake-codex" {
		t.Fatalf("renamed loser runner = %q want fake-codex", got)
	}

	be, _ := os.ReadFile(filepath.Join(bare, "events.jsonl"))
	le, _ := os.ReadFile(filepath.Join(loser, "events.jsonl"))
	if !strings.Contains(string(be), "from-opencode") {
		t.Fatalf("bare events want from-opencode, got %s", be)
	}
	if !strings.Contains(string(le), "from-codex") {
		t.Fatalf("loser events want from-codex, got %s", le)
	}

	// old nested paths gone
	assertPathMissing(t, filepath.Join(req.Home, "sessions", "fake-codex", "shared"))
	assertPathMissing(t, filepath.Join(req.Home, "sessions", "fake-opencode", "shared"))

	if v := layoutVersion(t, req.Home); v != 2 {
		t.Fatalf(".layout version = %d want 2", v)
	}

	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if !strings.Contains(combined, "shared__fake-codex") &&
		!strings.Contains(combined, "rename") &&
		!strings.Contains(combined, "collision") {
		t.Fatalf("expected rename/collision mention in report, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
