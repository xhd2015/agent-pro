# Scenario

**Feature**: per-tool handler time avg/med/min/max from events.jsonl duration_ms

```
# summary + events.jsonl with known tool_completed durations
writeGrokSessionOpts -> writeEventsJSONL -> sessions.Stats

# ToolStat aggregates duration_ms by tool_name; outcome drives Success/Error
```

## Preconditions

- Session has `summary.json` and `events.jsonl` (no signals, no updates).
- `read_file` has three successful completions: 10, 20, 30 ms.
- `bash` has one error completion: 50 ms.
- Corresponding `tool_started` lines exist for ToolCalls fallback.

## Steps

1. Write a minimal session summary.
2. Write `events.jsonl` with tool_started/tool_completed pairs as above.
3. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const eventsToolAvgSessionID = "019f283b-cccc-7ccc-cccc-cccccccccccc"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = eventsToolAvgSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, eventsToolAvgSessionID,
		"2026-07-03T14:10:00.000Z",
		"/tmp/grok-stats-events-tools",
		"Tool duration session",
		grokSessionOpts{
			NumMessages:     6,
			NumChatMessages: 3,
		})
	writeEventsJSONL(t, sessionDirOf(summaryPath), []map[string]any{
		toolStarted("read_file"),
		toolCompleted("read_file", 10, "success"),
		toolStarted("read_file"),
		toolCompleted("read_file", 20, "success"),
		toolStarted("read_file"),
		toolCompleted("read_file", 30, "success"),
		toolStarted("bash"),
		toolCompleted("bash", 50, "error"),
	})
	return nil
}
```
