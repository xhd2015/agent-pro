## Expected

- `FormatStatsTextOpts` output includes identity lines for Session, Title, Model, Agent.
- Section headers (stable for asserts):
  - `Counts`
  - `Latency`
  - `Tool handler time`
  - `Background tasks`
  - `Subagents`
  - `Sources`
- Count lines include Turns and Tool calls labels.
- Tool section mentions `read_file`.
- Latency lines have Duration / Avg response / Avg time-to-first labels
  (values may be pretty: `2m`, `1.5s`, `400ms` — not required raw `120s`).
- Sources line reflects used files (summary, signals, events, updates).

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	out := resp.Output
	if out == "" {
		t.Fatal("FormatStatsTextOpts output is empty")
	}

	assertContains(t, out, "Session: "+req.SessionID)
	assertContains(t, out, "Title:")
	assertContains(t, out, "Format stats text")
	assertContains(t, out, "Model:")
	assertContains(t, out, "grok-composer-2.5-fast")
	assertContains(t, out, "Agent:")
	assertContains(t, out, "cursor")

	assertContains(t, out, "Counts")
	assertContains(t, out, "Turns:")
	assertContains(t, out, "Tool calls:")
	assertContains(t, out, "Thinking blocks:")

	assertContains(t, out, "Latency")
	assertContains(t, out, "Duration:")
	assertContains(t, out, "Avg response:")
	assertContains(t, out, "Avg time-to-first:")

	assertContains(t, out, "Tool handler time")
	assertContains(t, out, "read_file")

	assertContains(t, out, "Background tasks")
	assertContains(t, out, "Subagents")
	assertContains(t, out, "Sources")
	assertContains(t, out, "summary")
	assertContains(t, out, "signals")
	assertContains(t, out, "events")
	assertContains(t, out, "updates")
}
```
