---
label: grok
explanation: Requires real grok CLI on PATH; for design verification and debugging.
---

## Expected

- Exit code 0.
- Captured output is non-empty: stdout has substantive text **or**
  `sessions/grok-tty/.../events.jsonl` has at least one assistant/message event
  with non-trivial content (not whitespace-only).

## Exit Code

0

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	stdout := strings.TrimSpace(resp.Stdout)
	if len(stdout) > 0 {
		return
	}
	_, lines := findGrokTTYEventsJSONL(t, req.Home)
	if len(lines) == 0 {
		t.Fatalf("expected non-empty stdout or events; stdout:\n%s", resp.Stdout)
	}
	for _, line := range lines {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		if typ, _ := ev["type"].(string); typ == "message" || typ == "action_message" {
			if msg, _ := ev["message"].(string); len(strings.TrimSpace(msg)) > 3 {
				return
			}
			if content, _ := ev["content"].(string); len(strings.TrimSpace(content)) > 3 {
				return
			}
		}
	}
	t.Fatalf("expected substantive assistant output in stdout or events; stdout:\n%s\nevents:\n%s",
		resp.Stdout, strings.Join(lines, "\n"))
}
```