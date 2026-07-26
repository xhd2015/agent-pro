# Scenario

**Feature**: ColorMode always applies ANSI (dim labels, red failures, green yes)

```
# same color-worthy fixture as color-never → FormatStatsTextOpts(ColorMode=always)
writeGrokSessionOpts + sidecars -> Stats -> FormatStatsTextOpts

# output contains CSI escapes (dim and/or green/red)
```

## Preconditions

- Session has tool ERROR counts and source `yes` marks so coloring is applicable.
- `req.ColorMode = "always"` forces color even when stdout is non-TTY.

## Steps

1. Write session with tool error outcomes, non-zero ToolFailed/Errors, full Sources.
2. Set `req.ColorMode = "always"`.
3. Set `req.SessionID`.

```go
import "testing"

const formatColorAlwaysSessionID = "019f283b-8888-7888-8888-888888888888"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = formatColorAlwaysSessionID
	req.ColorMode = "always"
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatColorAlwaysSessionID,
		"2026-07-03T14:59:00.000Z",
		"/tmp/grok-stats-color-always",
		"Color always session",
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
