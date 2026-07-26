---
label: e2e
---

## Expected

- Command succeeds (exit 0).
- JSON output has `"home"` field containing `$HOME/.agent-hub` (the user's home directory, not a relative `.agent-hub`).
- The `running` field is `false` (no daemon started).

```go
import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(resp.Stdout), &obj); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, resp.Stdout)
	}

	home, _ := obj["home"].(string)
	if home == "" {
		t.Fatalf("missing 'home' in output: %s", resp.Stdout)
	}

	expectedHome := filepath.Join(req.UserHomeDir, ".agent-hub")
	if home != expectedHome {
		t.Fatalf("expected home=%q, got %q", expectedHome, home)
	}

	running, _ := obj["running"].(bool)
	if running {
		t.Fatal("expected running=false")
	}
}
```
