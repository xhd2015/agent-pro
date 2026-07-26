# Scenario

**Feature**: session list prints durable map rows

```
# Caller lists local sessions (no Slack network)
Caller -> slack-msg session list [--json] [--limit N]
  -> load ~/.agent-pro/slack-local-bot/sessions.json
  -> sort updated_at desc
  -> human table or --json {"sessions":[...]}
```

## Preconditions

- Action is `list` as second arg.
- Isolated `HomeDir`; clear Slack env.
- No live Slack / tokens required.

## Steps

1. Isolate home under WorkDir.
2. Leaves seed `sessions.json` and set list flags.

## Context

- Human columns: `SESSION_ID`, `CHANNEL`, `DIR`, `UPDATED`, `PREVIEW`.
- Right-pad each column to max(header, cell) width; join with two spaces.
- Empty `dir` → human `-`; JSON `""`.
- JSON rows emit both `session_id` and `agent_session_id` (equal today).

```go
import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	sessionListNewerID = "slack-channel-C01ABCDEFF0"
	sessionListOlderID = "slack-dm-U02XYZABCDEF"
)

func sessionListFixtureEntries() []sessionMapEntry {
	return []sessionMapEntry{
		{
			SessionID:          sessionListOlderID,
			ChannelID:          "D02XYZABCDEF",
			ThreadTS:           "",
			ConfigPath:         "",
			Dir:                "",
			Kind:               "dm",
			ReplyMode:          "channel",
			LastMessagePreview: "ping",
			CreatedAt:          "2026-07-11T00:00:00Z",
			UpdatedAt:          "2026-07-12T01:00:00Z",
		},
		{
			SessionID:          sessionListNewerID,
			ChannelID:          "C01ABCDEFF0",
			ThreadTS:           "1710000000.000100",
			ConfigPath:         "/tmp/cfg.json",
			Dir:                "",
			Kind:               "channel",
			ReplyMode:          "channel",
			LastMessagePreview: "hello from slack",
			CreatedAt:          "2026-07-12T00:00:00Z",
			UpdatedAt:          "2026-07-13T07:46:00Z",
		},
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	if err := isolateHome(t, req); err != nil {
		return err
	}
	return nil
}

// formatSessionListHuman builds the locked grok-style table for asserts.
// Columns: SESSION_ID, CHANNEL, DIR, UPDATED, PREVIEW.
// Empty dir cell should already be "-" when passed in.
func formatSessionListHuman(rows [][]string) string {
	headers := []string{"SESSION_ID", "CHANNEL", "DIR", "UPDATED", "PREVIEW"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := 0; i < len(headers) && i < len(row); i++ {
			if n := len(row[i]); n > widths[i] {
				widths[i] = n
			}
		}
	}
	var b strings.Builder
	writeRow := func(cols []string) {
		for i := 0; i < len(headers); i++ {
			cell := ""
			if i < len(cols) {
				cell = cols[i]
			}
			if i > 0 {
				b.WriteString("  ")
			}
			if i == len(headers)-1 {
				b.WriteString(cell)
			} else {
				fmt.Fprintf(&b, "%-*s", widths[i], cell)
			}
		}
		b.WriteByte('\n')
	}
	writeRow(headers)
	for _, row := range rows {
		writeRow(row)
	}
	return b.String()
}
```
