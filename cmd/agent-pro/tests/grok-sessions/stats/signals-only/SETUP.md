# Scenario

**Feature**: stats fills counts and latency from signals when events/updates are missing

```
# summary.json + signals.json only (no events.jsonl, no updates.jsonl)
writeGrokSessionOpts -> writeSignalsJSON -> sessions.Stats

# Sources.Events/Updates false; Warnings mention missing optional files
# counts still succeed from signals
```

## Preconditions

- Session has `summary.json` and rich `signals.json`.
- `events.jsonl` and `updates.jsonl` are absent.
- Stats must still succeed (no hard fail on missing optional files).

## Steps

1. Write a session with model and agent identity.
2. Write `signals.json` with turn/tool/latency counters.
3. Do **not** write events or updates.
4. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const signalsOnlySessionID = "019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = signalsOnlySessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, signalsOnlySessionID,
		"2026-07-03T14:00:00.000Z",
		"/tmp/grok-stats-signals-only",
		"Signals only session",
		grokSessionOpts{
			NumMessages:     10,
			NumChatMessages: 4,
			CurrentModelID:  "grok-4",
			AgentName:       "default",
		})
	writeSignalsJSON(t, sessionDirOf(summaryPath), defaultStatsSignals())
	return nil
}
```
