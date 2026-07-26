## Expected

- Exit code 0.
- Stderr empty.
- Stdout is the add-missing-scope guideline and ends with trailing `\n`.
- Body must mention (contains, case-sensitive where noted):
  - Cannot add scopes to an existing token from this CLI (or equivalent)
  - `api.slack.com/apps`
  - OAuth & Permissions / Bot Token Scopes (or clear equivalent path)
  - `groups:read`
  - Reinstall (to workspace)
  - `botToken` and/or `SLACK_BOT_TOKEN` / config update
  - Retry the command
- Optional but preferred: note that official `slack` CLI still needs reinstall.

## Exit Code

0

```go
import (
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
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("topic stdout must end with trailing newline, got %q", resp.Stdout)
	}
	out := resp.Stdout
	low := strings.ToLower(out)

	// Cannot add scopes from this CLI.
	if !strings.Contains(low, "cannot") && !strings.Contains(low, "can't") && !strings.Contains(low, "not able") {
		t.Fatalf("topic should say scopes cannot be added from this CLI:\n%s", out)
	}
	if !strings.Contains(low, "scope") {
		t.Fatalf("topic missing scope language:\n%s", out)
	}
	if !strings.Contains(low, "api.slack.com/apps") {
		t.Fatalf("topic missing api.slack.com/apps:\n%s", out)
	}
	if !strings.Contains(low, "oauth") {
		t.Fatalf("topic missing OAuth & Permissions path:\n%s", out)
	}
	if !strings.Contains(low, "bot token") && !strings.Contains(out, "Bot Token Scopes") {
		t.Fatalf("topic missing Bot Token Scopes step:\n%s", out)
	}
	if !strings.Contains(out, "groups:read") {
		t.Fatalf("topic missing groups:read example:\n%s", out)
	}
	if !strings.Contains(low, "reinstall") {
		t.Fatalf("topic missing reinstall to workspace:\n%s", out)
	}
	if !strings.Contains(out, "botToken") && !strings.Contains(out, "SLACK_BOT_TOKEN") && !strings.Contains(low, "config") {
		t.Fatalf("topic missing botToken / SLACK_BOT_TOKEN / config update:\n%s", out)
	}
	if !strings.Contains(low, "retry") {
		t.Fatalf("topic missing retry command guidance:\n%s", out)
	}
}
```
