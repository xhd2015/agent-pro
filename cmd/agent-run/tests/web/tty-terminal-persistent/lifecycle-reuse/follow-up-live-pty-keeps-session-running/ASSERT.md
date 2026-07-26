---
label: e2e
---

## Expected

- Follow-up POST returns HTTP 202.
- The prompt reaches the mapped PTY.
- The web session status remains `running` shortly after dispatch, because
  writing bytes to the terminal is not the same as receiving a completed agent
  response.

## Side Effects

- The fake ptywrap records backend terminal input.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until live-PTY follow-up dispatch keeps the web
  turn running instead of marking it finished immediately.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.FollowUpStatus != 202 {
		t.Fatalf("follow-up status=%d body=%s", resp.FollowUpStatus, resp.FollowUpBody)
	}
	if req.PTYInputSeen == nil || !strings.Contains(*req.PTYInputSeen, req.FollowUpPrompt) {
		seen := ""
		if req.PTYInputSeen != nil {
			seen = *req.PTYInputSeen
		}
		t.Fatalf("follow-up prompt did not reach mapped PTY; input=%q", seen)
	}
	if resp.FollowUpSessionStatus != 200 {
		t.Fatalf("follow-up session status HTTP=%d body=%s", resp.FollowUpSessionStatus, resp.FollowUpSessionBody)
	}
	obj := decodeJSONBody(t, resp.FollowUpSessionBody)
	session, _ := obj["session"].(map[string]any)
	status, _ := session["status"].(string)
	if status != "running" {
		t.Fatalf("live-PTY follow-up should keep session running until terminal response is observed; status=%q body=%s input=%q", status, resp.FollowUpSessionBody, *req.PTYInputSeen)
	}
}
```
