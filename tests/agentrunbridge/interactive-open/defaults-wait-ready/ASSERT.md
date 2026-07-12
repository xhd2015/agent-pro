## Expected

- No API error.
- Launch argv matches open profile with session, grok-tty, auto-send, new-terminal, open, prompt.
- At least one status poll.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	want := []string{
		"run",
		"--session-id=sess-io-default",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
		"--open",
		"--",
		"interactive hello",
	}
	assertArgsEqual(t, resp.Args, want)
	if resp.StatusPollCalls < 1 {
		t.Fatal("expected wait-ready status poll")
	}
	if resp.LaunchCalls < 1 {
		t.Fatal("expected launch")
	}
}
```
