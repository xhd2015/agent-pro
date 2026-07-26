# Scenario

**Feature**: subagent duration aggregate from subagent_finished updates

```
# updates.jsonl subagent_finished with duration_ms
writeGrokSessionOpts -> writeUpdatesJSONL -> sessions.Stats

# Subagents Count / AvgMs / MaxMs from duration_ms
```

## Preconditions

- Two `subagent_finished` updates:
  - duration_ms = 1000, status completed
  - duration_ms = 3000, status completed
- Expected: Count=2, AvgMs=2000, MaxMs=3000.

## Steps

1. Write a minimal session summary.
2. Write `updates.jsonl` with the two subagent_finished lines.
3. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const subagentDurationSessionID = "019f283b-2222-7222-2222-222222222222"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = subagentDurationSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, subagentDurationSessionID,
		"2026-07-03T14:40:00.000Z",
		"/tmp/grok-stats-subagent",
		"Subagent duration session",
		grokSessionOpts{
			NumMessages:     5,
			NumChatMessages: 3,
		})
	writeUpdatesJSONL(t, sessionDirOf(summaryPath), []map[string]any{
		subagentFinished(1000, "completed"),
		subagentFinished(3000, "completed"),
	})
	return nil
}
```
