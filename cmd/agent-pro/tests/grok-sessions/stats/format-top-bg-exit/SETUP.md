# Scenario

**Feature**: Top background tasks EXIT column shows task_snapshot.exit_code

```
# task_completed with exit_code 1 and a short command
writeUpdatesJSONL -> FormatStatsTextOpts

# Top background tasks header has EXIT; row shows 1
```

## Preconditions

- One `task_completed` with:
  - `task_snapshot.command` = `false-cmd-exit-one`
  - `task_snapshot.exit_code` = **1**
  - wall clock 5s (start 1000 → end 1005)
- A second task with exit_code **0** and command `ok-cmd-exit-zero` so the
  table has two distinct EXIT values (optional but useful).
- Header must include `EXIT`.

## Steps

1. Write session summary.
2. Write updates with two `taskCompletedCmdExit` lines (exit 1 and 0).
3. Set `req.SessionID`.

```go
import "testing"

const formatTopBgExitSessionID = "019f283b-7002-7702-7702-770277027702"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = formatTopBgExitSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatTopBgExitSessionID,
		"2026-07-03T14:58:00.000Z",
		"/tmp/grok-stats-top-bg-exit",
		"Top bg exit column",
		grokSessionOpts{
			NumMessages:     3,
			NumChatMessages: 1,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	writeUpdatesJSONL(t, sessionDirOf(summaryPath), []map[string]any{
		taskCompletedCmdExit(1000, 1005, "false-cmd-exit-one", 1),
		taskCompletedCmdExit(2000, 2003, "ok-cmd-exit-zero", 0),
	})
	return nil
}
```
