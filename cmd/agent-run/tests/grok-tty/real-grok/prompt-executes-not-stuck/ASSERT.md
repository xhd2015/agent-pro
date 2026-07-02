---
label: grok
explanation: Requires real grok CLI on PATH; reproduces prompt-not-submitted bug with run ls.
---

## Expected

- Exit code 0.
- Grok processed the submitted prompt: assistant capture is not **only** the echoed
  user prompt `run ls` (the bug left stdout as `💬 run ls` with no grok response).
- Prefer evidence of command output (`total`, `drwx`) when grok runs `ls` via tools.
- Must not fail with `grok TUI banner not detected`.

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
	if strings.Contains(strings.ToLower(resp.Stderr), "banner not detected") {
		t.Fatalf("grok banner not detected:\n%s", resp.Stderr)
	}
	assertSuccess(t, resp)
	combined := resp.Stdout + "\n" + resp.Stderr
	_, lines := findGrokTTYEventsJSONL(t, req.Home)
	blob := strings.Join(lines, "\n") + "\n" + combined
	lower := strings.ToLower(blob)
	onlyEcho := strings.TrimSpace(resp.Stdout) == "💬 run ls\n[done]" ||
		strings.TrimSpace(resp.Stdout) == "💬 run ls\n[done] "
	for _, line := range lines {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		if typ, _ := ev["type"].(string); typ != "message" {
			continue
		}
		if role, _ := ev["role"].(string); role != "assistant" {
			continue
		}
		text, _ := ev["text"].(string)
		if strings.TrimSpace(text) != "" && !strings.EqualFold(strings.TrimSpace(text), "run ls") {
			onlyEcho = false
			break
		}
	}
	hasListing := strings.Contains(lower, "total ") || strings.Contains(lower, "drwx")
	if onlyEcho && !hasListing {
		t.Fatalf("prompt likely not submitted (only echoed run ls); stdout:\n%s\nevents+stderr blob:\n%s",
			resp.Stdout, blob)
	}
}
```