# Scenario

**Feature**: tool handler table sorts by N desc and shows NAME/N header columns

```
# events: read_file n=3, bash n=1 → table rows read_file before bash
writeGrokSessionOpts + writeEventsJSONL -> Stats -> FormatStatsTextOpts

# "Tool handler time" section has header with NAME and N; higher N first
```

## Preconditions

- Session has summary + events only (same duration mix as events-tool-avg).
- `read_file` Count=3, `bash` Count=1 → sorted by N: read_file above bash.
- Alphabetical order alone would put `bash` first — test requires N-desc order.

## Steps

1. Write minimal summary.
2. Write events with three successful read_file completions and one bash error.
3. Set `req.SessionID`.

```go
import "testing"

const formatToolTableSortSessionID = "019f283b-5555-7555-5555-555555555555"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = formatToolTableSortSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatToolTableSortSessionID,
		"2026-07-03T14:56:00.000Z",
		"/tmp/grok-stats-tool-sort",
		"Tool table sort session",
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
