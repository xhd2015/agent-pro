# Scenario

**Feature**: FormatStatsTextOpts prints pretty human durations on latency lines

```
# defaultStatsSignals: sessionDurationSeconds=120, avgResponseTimeMs=1500,
# avgTimeToFirstTokenMs=400
writeGrokSessionOpts + writeSignalsJSON -> Stats -> FormatStatsTextOpts

# Latency section shows 2m / 1.5s / 400ms (not raw 120s / 1500ms / 400ms only)
```

## Preconditions

- Session has summary + signals with `defaultStatsSignals` values:
  - 120 s session → pretty `2m`
  - 1500 ms avg response → pretty `1.5s`
  - 400 ms TTFT → pretty `400ms`
- ColorMode default never; TopN irrelevant for latency lines.

## Steps

1. Write a session with model/agent and `defaultStatsSignals`.
2. Set `req.SessionID` to the fixture UUID.
3. Leave ColorMode empty (Run → never).

```go
import "testing"

const formatPrettyDurationSessionID = "019f283b-4444-7444-4444-444444444444"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = formatPrettyDurationSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatPrettyDurationSessionID,
		"2026-07-03T14:55:00.000Z",
		"/tmp/grok-stats-pretty-dur",
		"Pretty duration session",
		grokSessionOpts{
			NumMessages:     8,
			NumChatMessages: 4,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	writeSignalsJSON(t, sessionDirOf(summaryPath), defaultStatsSignals())
	return nil
}
```
