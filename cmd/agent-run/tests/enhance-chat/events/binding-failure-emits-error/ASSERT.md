## Expected

- `events.jsonl` contains `think` with text `Resolve session id...`.
- `events.jsonl` contains `error` whose text starts with `Cannot resolve session id:`.
- No assistant `message` row with scrollback fallback text (`hi` or `Response:`).
- Session still reaches `finished` (error is surfaced in events, not a hung run).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !eventsHaveThinkText(resp.EventsParsed, resolveThinkText) {
		t.Fatalf("events.jsonl missing think %q; events=%v", resolveThinkText, resp.EventsParsed)
	}
	if !eventsHaveErrorPrefix(resp.EventsParsed, resolveErrorPrefix) {
		t.Fatalf("events.jsonl missing error prefix %q; events=%v", resolveErrorPrefix, resp.EventsParsed)
	}
	if eventsHaveAssistantMessage(resp.EventsParsed) {
		t.Fatalf("expected no assistant message fallback on bind failure; events=%v", resp.EventsParsed)
	}
	for _, ev := range resp.EventsParsed {
		text, _ := ev["text"].(string)
		trimmed := strings.TrimSpace(text)
		if trimmed == "hi" || strings.Contains(text, "Response: hi") {
			t.Fatalf("scrollback fallback text leaked into events.jsonl:\n%s", joinLines(resp.EventsFileLines))
		}
	}
}

func joinLines(lines []string) string {
	out := ""
	for _, line := range lines {
		out += line + "\n"
	}
	return out
}
```