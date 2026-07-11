## Expected Output

After attach returns, stderr includes (order may place terminal id first):

```text
grok-tty: <terminal-or-session-id>
grok-tty: grok session 550e8400-e29b-41d4-a716-446655440700
grok-tty: grok updates <path>/updates.jsonl
```

## Expected

- Exit code 0.
- Stderr contains `grok-tty: grok session` with the seeded UUID.
- Stderr contains `grok-tty: grok updates` and path ending in `updates.jsonl`.
- Some `sessions/grok-tty/*/meta.json` has `runner_session_id` equal to the UUID.

## Exit Code

0

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stderr, "grok-tty: grok session "+openBindGrokUUID)
	assertContains(t, resp.Stderr, "grok-tty: grok updates")
	assertContains(t, resp.Stderr, "updates.jsonl")

	// Persist check: any grok-tty session meta under home.
	found := false
	root := filepath.Join(req.Home, "sessions", "grok-tty")
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
		if id, _ := meta["runner_session_id"].(string); id == openBindGrokUUID {
			found = true
		}
		return nil
	})
	if !found {
		t.Fatalf("no meta.json with runner_session_id=%q under %s\nstderr:\n%s", openBindGrokUUID, root, resp.Stderr)
	}
	_ = strings.TrimSpace
}
```
