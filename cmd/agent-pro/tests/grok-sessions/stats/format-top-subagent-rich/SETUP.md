# Scenario

**Feature**: Top subagents shows spawn description, type, tools, turns (join)

```
# subagent_spawned then subagent_finished joined by subagent_id
writeUpdatesJSONL -> FormatStatsTextOpts

# Top subagents: STATUS TYPE TOOLS TURNS DESC — not UUID-only label
```

## Preconditions

- Spawn:
  - `subagent_id` = `019f283b-sa01-7a01-a001-a001a001a001`
  - `description` = `[designer] design rich top subagent tables`
  - `type` = `general-purpose`
  - `model` = `grok-composer-2.5-fast`
- Finish (same id):
  - `duration_ms` = 900000 (15m)
  - `status` = `completed`
  - `tool_calls` = 99
  - `turns` = 1
  - `tokens_used` = 12000
- No description on finish (must come from spawn join).
- Table headers include STATUS, TYPE, TOOLS, TURNS, DESC.

## Steps

1. Write session summary.
2. Write updates: `subagentSpawned` then `subagentFinishedFull`.
3. Set `req.SessionID`.

```go
import "testing"

const formatTopSubRichSessionID = "019f283b-7003-7703-7703-770377037703"
const formatTopSubRichSubID = "019f283b-sa01-7a01-a001-a001a001a001"
const formatTopSubRichDesc = "[designer] design rich top subagent tables"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = formatTopSubRichSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatTopSubRichSessionID,
		"2026-07-03T14:59:00.000Z",
		"/tmp/grok-stats-top-sub-rich",
		"Top subagent rich",
		grokSessionOpts{
			NumMessages:     6,
			NumChatMessages: 3,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	writeUpdatesJSONL(t, sessionDirOf(summaryPath), []map[string]any{
		subagentSpawned(formatTopSubRichSubID, formatTopSubRichDesc, "general-purpose", "grok-composer-2.5-fast"),
		subagentFinishedFull(formatTopSubRichSubID, 900000, "completed", 99, 1, 12000),
	})
	return nil
}
```
