---
label: unit
explanation: session list human header SESSION_ID + padding; sort updated_at desc; empty dir -
---

## Expected Output

```
SESSION_ID                 CHANNEL       DIR  UPDATED               PREVIEW
slack-channel-C01ABCDEFF0  C01ABCDEFF0   -    2026-07-13T07:46:00Z  hello from slack
slack-dm-U02XYZABCDEF      D02XYZABCDEF  -    2026-07-12T01:00:00Z  ping
```

(Exact padding is computed by `formatSessionListHuman` in parent SETUP.)

## Expected

- Exit code 0.
- Header first label is `SESSION_ID`; columns CHANNEL, DIR, UPDATED, PREVIEW.
- Rows sorted by `updated_at` descending (channel session first).
- Empty dir shown as `-`.
- Column padding (right-pad + two-space join) matches harness helper.
- Trailing newline; stderr empty; no tokens in output.

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
	want := formatSessionListHuman([][]string{
		{sessionListNewerID, "C01ABCDEFF0", "-", "2026-07-13T07:46:00Z", "hello from slack"},
		{sessionListOlderID, "D02XYZABCDEF", "-", "2026-07-12T01:00:00Z", "ping"},
	})
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, resp.Stdout)
	}
	// Guard: first header token must be SESSION_ID (locked product decision).
	firstLine := strings.Split(strings.TrimSuffix(resp.Stdout, "\n"), "\n")[0]
	if !strings.HasPrefix(firstLine, "SESSION_ID") {
		t.Fatalf("first header label must be SESSION_ID, got %q", firstLine)
	}
	for _, tok := range []string{"xoxb-", "xoxp-", "xapp-"} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("list output must not contain tokens, found %q", tok)
		}
	}
}
```
