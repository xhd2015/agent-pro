# Scenario

**Feature**: stats returns full identity, counts, tools, and tasks from all sources

```
# summary.json + signals.json + events.jsonl + updates.jsonl under encoded cwd
writeGrokSessionOpts -> writeSignalsJSON / writeEventsJSONL / writeUpdatesJSONL
  -> sessions.Stats -> FormatStatsText(now)

# SessionStats filled from every source; Sources flags all true
```

## Preconditions

- Session has summary, signals, events, and updates.
- Signals carry turn/tool/latency counters.
- Events carry multiple tool_completed lines.
- Updates carry thought chunks, one background task, and one subagent finish.

## Steps

1. Write a session with model, agent, and message counts.
2. Write rich `signals.json` via `defaultStatsSignals`.
3. Write `events.jsonl` with tool_started/tool_completed for two tools.
4. Write `updates.jsonl` with two consecutive thought chunks, a non-thought,
   one more thought, one task_completed (5s wall), and one subagent_finished.
5. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const knownStatsSessionID = "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = knownStatsSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, knownStatsSessionID,
		"2026-07-03T13:00:00.000Z",
		"/tmp/grok-stats-known",
		"Analyse session stats",
		grokSessionOpts{
			NumMessages:     90,
			NumChatMessages: 40,
			CreatedAt:       "2026-07-03T10:00:00.000Z",
			UpdatedAt:       "2026-07-03T12:30:00.000Z",
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	dir := sessionDirOf(summaryPath)

	writeSignalsJSON(t, dir, defaultStatsSignals())

	writeEventsJSONL(t, dir, []map[string]any{
		toolStarted("read_file"),
		toolCompleted("read_file", 10, "success"),
		toolStarted("read_file"),
		toolCompleted("read_file", 20, "success"),
		toolStarted("read_file"),
		toolCompleted("read_file", 30, "success"),
		toolStarted("bash"),
		toolCompleted("bash", 100, "error"),
	})

	writeUpdatesJSONL(t, dir, []map[string]any{
		thoughtChunk("plan step one"),
		thoughtChunk("plan step two"),
		nonThoughtUpdate("agent_message_chunk"),
		thoughtChunk("after gap"),
		taskCompleted(1000, 1005),
		subagentFinished(5000, "completed"),
	})
	return nil
}
```
