# Scenario

**Feature**: list table output shows unified grok-shaped columns and relative times

```
# three sessions with time.updated offsets from fixed now
writeOpencodeSession x3 -> sessions.List -> FormatListTable(now)

# table includes SESSION ID, LAST ACTIVE, TITLE, MSGS, CWD with relative deltas
terminal table text
```

## Preconditions

- `req.Now` is fixed at `2026-07-03T15:00:00.000Z` by root Setup.
- Message files exist so MSGS column is non-zero for at least one session.

## Steps

1. Create session A active at `req.Now` → `just now`, with 2 messages.
2. Create session B active 5 minutes before `req.Now` → `5m ago`, with 1 message.
3. Create session C active 2 hours before `req.Now` → `2h ago`, no messages.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	now := req.Now.UTC()

	writeOpencodeSession(t, req.DataDir, "proj_table", "ses_table_alpha", "Alpha refactor",
		"/tmp/project-a", now)
	writeOpencodeMessage(t, req.DataDir, "ses_table_alpha", "msg_alpha_01", opencodeMessageOpts{InputTokens: 10, OutputTokens: 5})
	writeOpencodeMessage(t, req.DataDir, "ses_table_alpha", "msg_alpha_02", opencodeMessageOpts{InputTokens: 20, OutputTokens: 10})

	writeOpencodeSession(t, req.DataDir, "proj_table", "ses_table_beta", "Beta bugfix",
		"/tmp/project-b", now.Add(-5*time.Minute))
	writeOpencodeMessage(t, req.DataDir, "ses_table_beta", "msg_beta_01", opencodeMessageOpts{InputTokens: 5, OutputTokens: 2})

	writeOpencodeSession(t, req.DataDir, "proj_table", "ses_table_gamma", "Gamma cleanup",
		"/tmp/project-c", now.Add(-2*time.Hour))
	return nil
}
```