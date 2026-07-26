---
label: e2e
---

## Expected
- Valid JSONL event lines with recognizable structure.
- Events include reasoning, command_execution, and a final message.
- All event types have corresponding item types.

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

    // Verify valid JSONL
    lines := strings.Split(strings.TrimSpace(stdout), "\n")
    hasCommand := false
    for i, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        var event map[string]any
        if e := json.Unmarshal([]byte(line), &event); e != nil {
            t.Fatalf("line %d: invalid JSON: %v\n%s", i, e, line)
        }
        if item, ok := event["item"].(map[string]any); ok {
            if itemType, _ := item["type"].(string); itemType == "command_execution" {
                hasCommand = true
            }
        }
    }
    if !hasCommand {
        t.Fatalf("expected at least one command_execution item, got:\n%s", stdout)
    }

    // Verify old hardcoded fake data is NOT present (should use real execution)
    fakeBait := []string{
        "src/\n  main.go\n  utils.go",
        "On branch main\nnothing to commit",
        "github.com/xhd2015/agent-pro/...",
        "See /tmp/docs/guide.md",
        "DATABASE_URL=postgres://localhost/db",
    }
    for _, bait := range fakeBait {
        if strings.Contains(stdout, bait) {
            t.Fatalf("found old fake data (should use real execution): %q", bait)
        }
    }
}
```
