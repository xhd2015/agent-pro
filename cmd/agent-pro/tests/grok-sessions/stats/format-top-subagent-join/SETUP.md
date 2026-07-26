# Scenario

**Feature**: subagent_finished without spawn falls back to id; no crash

```
# finish-only updates.jsonl (no subagent_spawned)
writeUpdatesJSONL -> Stats / FormatStatsTextOpts

# Top subagents still renders; DESC falls back to short id; join miss is OK
```

## Preconditions

- Single `subagent_finished` with:
  - `subagent_id` = `019f283b-solo-7b0b-b00b-b00bb00bb00b`
  - `duration_ms` = 3000
  - `status` = `completed`
  - `tool_calls` = 2
  - `turns` = 1
  - **no** prior `subagent_spawned`
  - **no** `description` on finish
- Stats must succeed (no panic / error).
- DESC column (or label) falls back to the id (full or short prefix acceptable
  as long as a distinctive id fragment appears).

## Steps

1. Write session summary.
2. Write updates with only `subagentFinishedFull` (no spawn).
3. Set `req.SessionID`.

```go
import "testing"

const formatTopSubJoinSessionID = "019f283b-7004-7704-7704-770477047704"
const formatTopSubJoinSubID = "019f283b-solo-7b0b-b00b-b00bb00bb00b"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = formatTopSubJoinSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatTopSubJoinSessionID,
		"2026-07-03T15:00:00.000Z",
		"/tmp/grok-stats-top-sub-join",
		"Top subagent join miss",
		grokSessionOpts{
			NumMessages:     2,
			NumChatMessages: 1,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	writeUpdatesJSONL(t, sessionDirOf(summaryPath), []map[string]any{
		subagentFinishedFull(formatTopSubJoinSubID, 3000, "completed", 2, 1, 100),
	})
	return nil
}
```
