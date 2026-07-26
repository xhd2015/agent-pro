---
label: e2e
---

## Expected

- Two HTTP 200 responses on `POST /v1/messages`.
- Response #1 `content[]`: at least one `{type:"thinking"}` block with `preset:think:think-tool-message`; exactly one `{type:"tool_use"}` with `name` = `bash`; no `{type:"text"}` with message preset text.
- Response #2 `content[]`: `{type:"text"}` containing `preset:message:think-tool-message`; no `tool_use`.
- Agent-events: 3 lines — `think`, `tool_call`, `message`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
	}

	// Response #1: thinking + tool_use
	r1 := resp.Responses[0]
	if r1.StatusCode != 200 {
		t.Fatalf("response 1: expected 200, got %d\nbody: %s", r1.StatusCode, r1.Body)
	}
	blocks1 := anthropicContentBlocks(t, r1.Body)
	var thinkingCount, toolUseCount, textCount int
	for _, b := range blocks1 {
		typ, _ := b["type"].(string)
		switch typ {
		case "thinking":
			thinkingCount++
			thinking, _ := b["thinking"].(string)
			if thinking == "" {
				// some APIs use "text" subfield
				thinking, _ = b["text"].(string)
			}
			if thinking != "" {
				assertContains(t, thinking, "preset:think:think-tool-message")
			}
		case "tool_use":
			toolUseCount++
			name, _ := b["name"].(string)
			if name != "bash" {
				t.Fatalf("response 1: expected tool_use name=bash, got %q", name)
			}
		case "text":
			textCount++
			text, _ := b["text"].(string)
			if text != "" {
				assertNotContains(t, text, "preset:message:think-tool-message")
			}
		}
	}
	if thinkingCount < 1 {
		t.Fatalf("response 1: expected thinking block, got blocks: %v", blocks1)
	}
	if toolUseCount != 1 {
		t.Fatalf("response 1: expected exactly 1 tool_use, got %d", toolUseCount)
	}

	// Response #2: text only
	r2 := resp.Responses[1]
	if r2.StatusCode != 200 {
		t.Fatalf("response 2: expected 200, got %d\nbody: %s", r2.StatusCode, r2.Body)
	}
	blocks2 := anthropicContentBlocks(t, r2.Body)
	var textOnly bool
	for _, b := range blocks2 {
		typ, _ := b["type"].(string)
		if typ == "tool_use" {
			t.Fatalf("response 2: unexpected tool_use block: %v", b)
		}
		if typ == "text" {
			text, _ := b["text"].(string)
			if text != "" {
				assertContains(t, text, "preset:message:think-tool-message")
				textOnly = true
			}
		}
	}
	if !textOnly {
		t.Fatalf("response 2: expected text block with message preset text, got: %v", blocks2)
	}

	if len(resp.AgentEventsLines) != 3 {
		t.Fatalf("agent-events: want 3 lines, got %d\n%s", len(resp.AgentEventsLines), resp.AgentEventsContent)
	}
	events, parseErr := parseAgentEventMaps(resp.AgentEventsLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if events[0]["type"] != "think" || events[1]["type"] != "tool_call" || events[2]["type"] != "message" {
		t.Fatalf("want think, tool_call, message order, got %v, %v, %v",
			events[0]["type"], events[1]["type"], events[2]["type"])
	}
}
```