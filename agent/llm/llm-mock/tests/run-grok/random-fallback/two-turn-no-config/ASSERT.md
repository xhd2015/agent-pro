---
label: e2e
---

## Expected

- Fake grok exits 0 (all curls succeed; no `curl -sf` failure from HTTP 400).
- Combined stdout/stderr contains `R1=`, `R2=`, `R3=` with JSON bodies.
- None of the three responses contain `error.type` = `"no_match"`.
- Third response has generated assistant content (second user turn answered).

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	combined := resp.Stdout + resp.Stderr
	for _, label := range []string{"R1=", "R2=", "R3="} {
		if !strings.Contains(combined, label) {
			t.Fatalf("missing %s in output:\n%s", label, combined)
		}
	}
	assertNotContains(t, combined, `"type":"no_match"`)
	assertNotContains(t, combined, `"type": "no_match"`)

	// Parse R3 body and require a non-empty assistant message (not no_match).
	r3Line := ""
	for _, line := range strings.Split(combined, "\n") {
		if strings.HasPrefix(line, "R3=") {
			r3Line = strings.TrimPrefix(line, "R3=")
			break
		}
	}
	if r3Line == "" {
		t.Fatalf("R3: curl failed (empty body) — second user turn got HTTP 400 no_match after turn 1 exhausted genStream.\nR1/R2 succeeded; expected R3 JSON body.\noutput:\n%s", combined)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(r3Line), &obj); err != nil {
		t.Fatalf("R3 JSON parse: %v\nbody: %s", err, r3Line)
	}
	if errObj, ok := obj["error"].(map[string]any); ok {
		t.Fatalf("R3: expected 200 random fallback for second user turn, got error: %v", errObj)
	}
	choices, ok := obj["choices"].([]any)
	if !ok || len(choices) == 0 {
		t.Fatalf("R3: expected choices, got %v", obj)
	}
	msg := choices[0].(map[string]any)["message"].(map[string]any)
	content, _ := msg["content"].(string)
	if content == "" && msg["tool_calls"] == nil {
		t.Fatalf("R3: expected generated content for \"what's wrong with me?\", got empty message: %v", msg)
	}
}
```