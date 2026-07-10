---
label: unit
explanation: slacktest API error path; requires SLACK_API_URL hook
---

## Expected

- Exit code 1.
- Stdout contains `Sending to channel=INVALID_CHANNEL`.
- Stderr contains `send failed:`.
- No `OK ts=` in stdout.

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStdoutContains(t, resp, `Sending to channel=INVALID_CHANNEL: "test"`)
	assertStderrContains(t, resp, "send failed:")
	if strings.Contains(resp.Stdout, "OK ts=") {
		t.Fatalf("unexpected OK line in stdout:\n%s", resp.Stdout)
	}
}
```