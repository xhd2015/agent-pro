# Scenario

**Feature**: Top-N sections list tools by total handler time; TopN=0 hides tops

```
# three tools with distinct Count*AvgMs; rich bg/sub items
writeGrokSessionOpts + events + updates -> FormatStatsTextOpts(TopN=2)

# "Top tools by total handler time" with at most 2 numbered rows
# Top background / Top subagents use rich headers (EXIT/COMMAND, STATUS/…)
# secondary assert: FormatStatsTextOpts(TopN=0) has no Top section headers
```

## Preconditions

- Tools (for total time ranking `Count*AvgMs`):
  - `search_replace`: n=2 × 100ms → total 200
  - `read_file`: n=3 × 10ms → total 30
  - `bash`: n=1 × 50ms → total 50
  - With TopN=2: top two by total are `search_replace` then `bash` (or
    whichever implementer ranks by total desc).
- Background: one task with command `doctest test ./... long-label`, exit 0.
- Subagent: spawn description `explore stats tree` + type `explore`, finish
  with tools/turns; DESC must not be UUID-only when description is present.
- Request: `TopN=2`, `TopNSet=true`.

## Steps

1. Write session with signals + events for three tools + updates for bg/sub.
2. Set `req.TopN = 2`, `req.TopNSet = true`.
3. Set `req.SessionID`.

```go
import "testing"

const formatTopNSessionID = "019f283b-6666-7666-6666-666666666666"
const formatTopNSubID = "019f283b-6666-sa66-6666-666666666666"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = formatTopNSessionID
	req.TopN = 2
	req.TopNSet = true
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatTopNSessionID,
		"2026-07-03T14:57:00.000Z",
		"/tmp/grok-stats-top-n",
		"Top N session",
		grokSessionOpts{
			NumMessages:     10,
			NumChatMessages: 5,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	dir := sessionDirOf(summaryPath)
	writeSignalsJSON(t, dir, defaultStatsSignals())
	writeEventsJSONL(t, dir, []map[string]any{
		toolStarted("search_replace"),
		toolCompleted("search_replace", 100, "success"),
		toolStarted("search_replace"),
		toolCompleted("search_replace", 100, "success"),
		toolStarted("read_file"),
		toolCompleted("read_file", 10, "success"),
		toolStarted("read_file"),
		toolCompleted("read_file", 10, "success"),
		toolStarted("read_file"),
		toolCompleted("read_file", 10, "success"),
		toolStarted("bash"),
		toolCompleted("bash", 50, "success"),
	})
	writeUpdatesJSONL(t, dir, []map[string]any{
		taskCompletedCmdExit(1000, 1005, "doctest test ./... long-label", 0),
		subagentSpawned(formatTopNSubID, "explore stats tree", "explore", "grok-composer-2.5-fast"),
		subagentFinishedFull(formatTopNSubID, 5000, "completed", 3, 1, 500),
	})
	return nil
}
```
