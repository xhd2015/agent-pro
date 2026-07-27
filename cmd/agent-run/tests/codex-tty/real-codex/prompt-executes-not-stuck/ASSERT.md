---
label: codex
explanation: Requires real codex CLI on PATH; reproduces prompt-not-submitted bug with run ls.
---

## Expected

- Exit code 0.
- Codex processed the submitted prompt: assistant capture is not **only** the echoed
  user prompt `run ls` (the bug left stdout as `💬 run ls` with no codex response).
- A screen that only reports `Queued follow-up inputs` for `run ls` is a failure,
  because the prompt was queued before Codex was ready instead of processed.
- Prefer evidence of command output (`total`, `drwx`) when codex runs `ls` via tools.
- Must not fail with `codex TUI banner not detected`.

## Exit Code

0

```go
import (
	"encoding/json"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(resp.Stderr), "banner not detected") {
		t.Fatalf("codex banner not detected:\n%s", resp.Stderr)
	}
	assertSuccess(t, resp)
	combined := resp.Stdout + "\n" + resp.Stderr
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	blob := strings.Join(lines, "\n") + "\n" + combined
	lower := strings.ToLower(blob)
	processedEvidence := []string{
		"total ",
		"drwx",
		"requirement-design-codex-tty-model-loading-readiness.md",
		"cmd/agent-run",
		"pkgs/groktty",
		"assistant",
		"tool",
	}
	hasProcessingEvidence := false
	for _, marker := range processedEvidence {
		if strings.Contains(lower, marker) {
			hasProcessingEvidence = true
			break
		}
	}
	if strings.Contains(lower, "queued follow-up inputs") &&
		strings.Contains(lower, "run ls") &&
		!hasProcessingEvidence {
		t.Fatalf("prompt was queued but not processed; stdout:\n%s\nevents+stderr blob:\n%s",
			resp.Stdout, blob)
	}
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
	if onlyEcho && !hasProcessingEvidence {
		t.Fatalf("prompt likely not submitted (only echoed run ls); stdout:\n%s\nevents+stderr blob:\n%s",
			resp.Stdout, blob)
	}
}
```
