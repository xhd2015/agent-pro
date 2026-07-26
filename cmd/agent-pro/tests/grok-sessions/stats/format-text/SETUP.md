# Scenario

**Feature**: FormatStatsTextOpts renders stable section headers and key counts

```
# full fixture (signals + events + updates) → FormatStatsTextOpts(stats, opts)
writeGrokSessionOpts + sidecars -> sessions.Stats -> FormatStatsTextOpts

# human text includes Counts, Latency, Tool handler time, Background tasks,
# Subagents, Sources section markers (durations may be pretty-printed)
```

## Preconditions

- Session has enough data that every optional section is present.
- Asserts focus on stable section headers and a few concrete values, not full layout.
- Duration lines may use pretty forms (`2m`, `1.5s`); do not require raw `120s`/`1500ms`.

## Steps

1. Write a session with model/agent and full sidecars (signals, events, updates).
2. Events include one tool_completed for `read_file`.
3. Updates include one thought, one task_completed, one subagent_finished.
4. Set `req.SessionID` to the fixture UUID.
5. Leave ColorMode empty (never) so header asserts stay free of ANSI.

```go
import "testing"

const formatTextSessionID = "019f283b-3333-7333-3333-333333333333"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = formatTextSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatTextSessionID,
		"2026-07-03T14:50:00.000Z",
		"/tmp/grok-stats-format",
		"Format stats text",
		grokSessionOpts{
			NumMessages:     12,
			NumChatMessages: 6,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	dir := sessionDirOf(summaryPath)
	writeSignalsJSON(t, dir, defaultStatsSignals())
	writeEventsJSONL(t, dir, []map[string]any{
		toolStarted("read_file"),
		toolCompleted("read_file", 42, "success"),
	})
	writeUpdatesJSONL(t, dir, []map[string]any{
		thoughtChunk("reason about format"),
		taskCompleted(1000, 1002),
		subagentFinished(900, "completed"),
	})
	return nil
}
```
