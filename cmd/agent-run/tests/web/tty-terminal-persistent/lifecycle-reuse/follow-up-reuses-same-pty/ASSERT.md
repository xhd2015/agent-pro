---
label: e2e
---

## Expected

- Terminal status is available before follow-up.
- Follow-up POST returns HTTP 202.
- Terminal status after follow-up is still available through `session-1`.
- Registry ids remain exactly `["session-1"]`.

## Side Effects

- The session receives one queued/running follow-up user message.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until follow-up dispatch preserves and reuses
  terminal mapping.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.FirstHTTPStatus != 200 {
		t.Fatalf("first terminal status=%d body=%s", resp.FirstHTTPStatus, resp.FirstHTTPBody)
	}
	requireTerminalMappingAvailable(t, req, resp.FirstHTTPBody)
	requireSameStringSlice(t, req.RegistryIDsBefore, []string{req.TerminalSessionID})
	if resp.FollowUpStatus != 202 {
		t.Fatalf("follow-up status=%d body=%s", resp.FollowUpStatus, resp.FollowUpBody)
	}
	if resp.SecondHTTPStatus != 200 {
		t.Fatalf("second terminal status=%d body=%s", resp.SecondHTTPStatus, resp.SecondHTTPBody)
	}
	requireTerminalMappingAvailable(t, req, resp.SecondHTTPBody)
	requireSameStringSlice(t, resp.RegistryIDsAfter, []string{req.TerminalSessionID})
}
```
