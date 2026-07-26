# Scenario

**Feature**: background task wall-clock aggregate from task_completed updates

```
# updates.jsonl task_completed with start_time / end_time epoch fields
writeGrokSessionOpts -> writeUpdatesJSONL -> sessions.Stats

# BackgroundTasks Count / AvgMs / MaxMs from wall-clock milliseconds
```

## Preconditions

- Two `task_completed` updates:
  - start 1000 → end 1005  → 5000 ms
  - start 2000 → end 2010  → 10000 ms
- Expected: Count=2, AvgMs=7500, MaxMs=10000.

## Steps

1. Write a minimal session summary.
2. Write `updates.jsonl` with the two task_completed lines.
3. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const backgroundTaskSessionID = "019f283b-1111-7111-1111-111111111111"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = backgroundTaskSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, backgroundTaskSessionID,
		"2026-07-03T14:30:00.000Z",
		"/tmp/grok-stats-bg-task",
		"Background tasks session",
		grokSessionOpts{
			NumMessages:     3,
			NumChatMessages: 2,
		})
	writeUpdatesJSONL(t, sessionDirOf(summaryPath), []map[string]any{
		taskCompleted(1000, 1005),
		taskCompleted(2000, 2010),
	})
	return nil
}
```
