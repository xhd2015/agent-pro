---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Stderr (or combined output) reports grok session id not resolved for the session.
- No `sessions/grok-tty/*/meta.json` has a non-empty false `runner_session_id`
  that looks like a successful bind UUID (must not invent a bound id on failure).

## Exit Code

non-zero

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("exec error: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when bind fails after full wait; elapsed=%s\nstderr:\n%s\nstdout:\n%s",
			resp.Elapsed, resp.Stderr, resp.Stdout)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"grok session id not resolved",
		"session id not resolved",
		"not resolved for session",
		"not resolved",
	)

	// No false successful bind id under home.
	root := filepath.Join(req.Home, "sessions")
	_ = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Name() != "meta.json" {
			return nil
		}
		data, rErr := os.ReadFile(path)
		if rErr != nil {
			return nil
		}
		var meta map[string]any
		if json.Unmarshal(data, &meta) != nil {
			return nil
		}
		if id, _ := meta["runner_session_id"].(string); strings.TrimSpace(id) != "" {
			t.Fatalf("unexpected runner_session_id=%q after hard bind failure in %s\nstderr:\n%s", id, path, resp.Stderr)
		}
		return nil
	})
}
```
