---
label: unit
explanation: "bot --json document with config (none), identity fields, token_masked"
---

## Expected Output

```json
{"config":"(none)","kind":"bot","ok":true,"team":"SlackTest Team","team_id":"T024BE7LD","user":"Egon Spengler","user_id":"W012A3CDE","bot_id":"B0TESTBOTID","url":"https://localhost.localdomain/","token_masked":"xoxb-...oken"}
```

## Expected

- Exit code 0.
- Single JSON document on stdout (trailing newline via json.Encoder).
- Fields: `config`=`(none)`, `kind`=`bot`, `ok`=true, identity fields, `token_masked`.
- No raw full token in stdout.
- Stderr empty.

## Exit Code

0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, slackTestToken) {
		t.Fatalf("stdout must not contain raw bot token %q:\n%s", slackTestToken, resp.Stdout)
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
	var doc struct {
		Config      string `json:"config"`
		Kind        string `json:"kind"`
		Ok          bool   `json:"ok"`
		Team        string `json:"team"`
		TeamID      string `json:"team_id"`
		User        string `json:"user"`
		UserID      string `json:"user_id"`
		BotID       string `json:"bot_id"`
		URL         string `json:"url"`
		TokenMasked string `json:"token_masked"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	if doc.Config != "(none)" {
		t.Fatalf("config = %q, want (none)", doc.Config)
	}
	if doc.Kind != "bot" || !doc.Ok {
		t.Fatalf("kind/ok = %q/%v, want bot/true", doc.Kind, doc.Ok)
	}
	if doc.Team != slackTestTeamName || doc.TeamID != slackTestTeamID {
		t.Fatalf("team fields = %q/%q, want %q/%q", doc.Team, doc.TeamID, slackTestTeamName, slackTestTeamID)
	}
	if doc.User != slackTestUserName || doc.UserID != slackTestUserID {
		t.Fatalf("user fields = %q/%q, want %q/%q", doc.User, doc.UserID, slackTestUserName, slackTestUserID)
	}
	if doc.BotID != slackTestAuthBotID {
		t.Fatalf("bot_id = %q, want %q", doc.BotID, slackTestAuthBotID)
	}
	if doc.URL != slackTestAuthURL {
		t.Fatalf("url = %q, want %q", doc.URL, slackTestAuthURL)
	}
	if doc.TokenMasked != slackTestTokenMasked {
		t.Fatalf("token_masked = %q, want %q", doc.TokenMasked, slackTestTokenMasked)
	}
}
```
