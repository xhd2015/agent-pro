---
label: e2e
---

## Expected
- The generated events contain command_execution items with real output.
- The old hardcoded fake strings (e.g., `"src/\n  main.go\n  utils.go"`) are NOT present.
- Events are valid JSON with the expected structure (type, item).

```go
import (
    "encoding/json"
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    stdout := resp.Stdout
    if strings.TrimSpace(stdout) == "" {
        t.Fatal("expected non-empty stdout from generator")
    }

    // Verify it's valid JSONL
    lines := strings.Split(strings.TrimSpace(stdout), "\n")
    if len(lines) < 1 {
        t.Fatal("expected at least one JSON line")
    }
    for i, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        var event map[string]any
        if err := json.Unmarshal([]byte(line), &event); err != nil {
            t.Fatalf("line %d: invalid JSON: %v\n%s", i, err, line)
        }
    }

    // Verify old hardcoded fake data is NOT present (should use real execution now)
    fakeBait := []string{
        "src/\n  main.go\n  utils.go",
        "github.com/xhd2015/agent-pro/...",
        "On branch main\nnothing to commit",
    }
    for _, bait := range fakeBait {
        if strings.Contains(stdout, bait) {
            t.Fatalf("found old fake data in output (should use real execution): %q", bait)
        }
    }

    // Verify at least one command_execution event exists
    if !strings.Contains(stdout, `"command_execution"`) {
        t.Fatalf("expected command_execution items, got:\n%s", stdout)
    }
}
```
