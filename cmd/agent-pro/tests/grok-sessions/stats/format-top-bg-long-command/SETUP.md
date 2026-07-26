# Scenario

**Feature**: Top background COMMAND shows long commands fully up to 200 runes

```
# updates.jsonl task_completed with command length 120 and 220
writeUpdatesJSONL -> Stats -> FormatStatsTextOpts(TopN=5)

# Top background tasks table: COMMAND column; 120-rune cmd fully visible;
# 220-rune cmd display-truncated with …
```

## Preconditions

- Two background tasks with full commands on `task_snapshot.command`:
  - Mid: exactly **120** ASCII runes ending in marker `MID120_OK` (within
    100–150 range → must appear **without** truncation ellipsis from the
    old 60-cap or 80 store-cap).
  - Long: exactly **220** ASCII runes ending in marker `LONG220_TAIL` →
    full string must **not** appear; display truncates at **200 runes**
    then `…` (Unicode ellipsis).
- Exit codes optional (0 for both).
- `TopN` default (5) or set so both rows can appear when ordered by duration.
- Mid task duration longer so it ranks first if useful (not required).

## Steps

1. Write session summary under a unique UUID.
2. Write `updates.jsonl` with two `taskCompletedCmdExit` lines (120- and
   220-rune commands).
3. Set `req.SessionID`; leave TopN default (5).

```go
import (
	"strings"
	"testing"
)

const formatTopBgLongCmdSessionID = "019f283b-7001-7701-7701-770177017701"

// midBgCommand is 120 runes (ASCII); must appear fully in COMMAND column.
var midBgCommand = strings.Repeat("m", 111) + "MID120_OK" // 111+9=120

// longBgCommand is 220 runes; display must truncate (200 runes + …).
var longBgCommand = strings.Repeat("L", 208) + "LONG220_TAIL" // 208+12=220

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if n := len([]rune(midBgCommand)); n != 120 {
		t.Fatalf("midBgCommand rune len = %d, want 120", n)
	}
	if n := len([]rune(longBgCommand)); n != 220 {
		t.Fatalf("longBgCommand rune len = %d, want 220", n)
	}
	req.SessionID = formatTopBgLongCmdSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, formatTopBgLongCmdSessionID,
		"2026-07-03T14:57:00.000Z",
		"/tmp/grok-stats-top-bg-long",
		"Top bg long command",
		grokSessionOpts{
			NumMessages:     4,
			NumChatMessages: 2,
			CurrentModelID:  "grok-composer-2.5-fast",
			AgentName:       "cursor",
		})
	dir := sessionDirOf(summaryPath)
	// Mid (120) longer duration → ranks above long (220) if sorted by duration.
	writeUpdatesJSONL(t, dir, []map[string]any{
		taskCompletedCmdExit(1000, 1010, midBgCommand, 0),  // 10s
		taskCompletedCmdExit(2000, 2005, longBgCommand, 0), // 5s
	})
	return nil
}
```
