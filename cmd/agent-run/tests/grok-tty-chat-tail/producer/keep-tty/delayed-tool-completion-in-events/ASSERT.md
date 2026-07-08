## Expected

- `events.jsonl` contains assistant `message` with `CHAT_TAIL_ASSISTANT_MARKER`.
- `events.jsonl` contains completed `tool_call` (status completed or completed update reflected).
- `done` event appears **after** the assistant marker line (correct ordering).
- Not only the initial pending tool_call — delayed completion must be converted.

## Side Effects

- `updates.jsonl` receives scheduled completion lines while keep-tty run is active.

## Exit Code

0 (run may be killed after events satisfy probe)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.EventsFileLines) == 0 {
		t.Fatalf("events.jsonl empty; stderr:\n%s", resp.Stderr)
	}
	if !resp.HasAssistantMarker {
		t.Fatalf("events.jsonl missing assistant marker %q; events=%v stderr:\n%s",
			chatTailAssistantMarker, resp.EventsParsed, resp.Stderr)
	}
	if !resp.HasCompletedTool {
		t.Fatalf("events.jsonl missing completed tool_call; events=%v", resp.EventsParsed)
	}
	if !resp.DoneAfterAssistant {
		t.Fatalf("expected done after assistant marker; events=%v", resp.EventsParsed)
	}
}
```