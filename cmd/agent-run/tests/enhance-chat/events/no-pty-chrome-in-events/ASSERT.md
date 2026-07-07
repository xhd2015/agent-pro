## Expected

- `events.jsonl` does not contain PTY chrome substrings: `╭`, `Grok Build`,
  `Starting session`, `Shift+Tab`, `GROK_TTY_BANNER`.
- Bind failure still emits explicit `error` event (not silent drop).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if marker, found := eventsLinesContainAnyChrome(resp.EventsFileLines); found {
		t.Fatalf("PTY chrome substring %q leaked into events.jsonl:\n%s", marker, joinEventLines(resp.EventsFileLines))
	}
	if !eventsHaveErrorPrefix(resp.EventsParsed, resolveErrorPrefix) {
		t.Fatalf("expected bind failure error event; events=%v", resp.EventsParsed)
	}
}

func joinEventLines(lines []string) string {
	out := ""
	for _, line := range lines {
		out += line + "\n"
	}
	return out
}
```