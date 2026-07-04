# Scenario

**Feature**: info succeeds without signals.json and omits Tokens section

```
# summary.json only, no signals.json sidecar
writeGrokSessionOpts -> sessions.Info -> FormatInfoText(now)

# Files block present; Tokens section absent
terminal key-value text
```

## Preconditions

- Session directory has `summary.json` but no `signals.json`.
- Info must still succeed.

## Steps

1. Write a minimal session without `WriteSignals`.
2. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const noSignalsSessionID = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbbbb"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = noSignalsSessionID
	writeGrokSessionOpts(t, req.GrokHome, noSignalsSessionID,
		"2026-07-03T14:00:00.000Z",
		"/tmp/grok-no-signals",
		"Docs cleanup",
		grokSessionOpts{
			NumMessages:     5,
			NumChatMessages: 3,
		})
	return nil
}
```