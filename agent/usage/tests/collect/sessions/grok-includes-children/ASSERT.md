## Expected

- `total` is 4: the main session, its subagent child, the fork, and the session
  without timestamps. Sessions below another session's directory count, so a
  snapshot's total reflects all activity on the account.
- The malformed summary is skipped: it is not a session we can describe.
- `oldest` is the earliest `created_at` (2026-09-01T08:00:00Z) and `newest` the
  latest `updated_at` (2026-09-03T09:00:00Z).
- A session with no timestamps still counts towards `total` but does not move
  either bound.
- Providers with no session tree report total 0 and no error, so the block never
  fails a collect.

## Side Effects

- None.

## Errors

- None: an unparsable summary is skipped, not reported as a walk error.

```go
import (
"testing"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

grok := resp.Record(t, usage.Grok)
AssertSessions(t, grok, 4, "2026-09-01T08:00:00Z", "2026-09-03T09:00:00Z")

for _, id := range []usage.ProviderID{usage.Codex, usage.CommandCode} {
rec := resp.Record(t, id)
AssertSessions(t, rec, 0, "", "")
}
}
```
