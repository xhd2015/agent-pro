# Scenario

**Feature**: ColorMode never yields plain stats text without ANSI

```
# full fixture with tool errors and sources → FormatStatsTextOpts(ColorMode=never)
writeGrokSessionOpts + sidecars -> Stats -> FormatStatsTextOpts

# output contains zero CSI escape sequences
```

## Preconditions

- Session has tools with at least one error and all Sources yes (color-worthy
  data present so "never" is meaningful).
- `req.ColorMode = "never"`.

## Steps

1. Write session with signals (toolFailureCount>0), events with one error tool,
   and updates so Sources are fully true.
2. Set `req.ColorMode = "never"`.
3. Set `req.SessionID`.

```go
import "testing"

const formatColorNeverSessionID = "019f283b-7777-7777-7777-777777777777"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = formatColorNeverSessionID
	req.ColorMode = "never"
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatColorNeverSessionID,
		"2026-07-03T14:58:00.000Z",
		"/tmp/grok-stats-color-never",
		"Color never session",
		grokSessionOpts{
			NumMessages:     8,
			NumChatMessages: 4,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	dir := sessionDirOf(summaryPath)
	sig := defaultStatsSignals()
	sig["toolFailureCount"] = 1
	sig["errorCount"] = 2
	writeSignalsJSON(t, dir, sig)
	writeEventsJSONL(t, dir, []map[string]any{
		toolStarted("bash"),
		toolCompleted("bash", 50, "error"),
		toolStarted("read_file"),
		toolCompleted("read_file", 5, "success"),
	})
	writeUpdatesJSONL(t, dir, []map[string]any{
		thoughtChunk("note"),
		taskCompleted(1000, 1002),
		subagentFinished(100, "completed"),
	})
	return nil
}
```
