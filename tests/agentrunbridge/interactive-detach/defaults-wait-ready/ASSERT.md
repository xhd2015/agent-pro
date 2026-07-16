## Expected

- No API error.
- Launch argv matches detach profile with session, grok-tty, auto-send, detach,
  `--`, prompt.
- No `--open`, no `--new-terminal`.
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
		"--session-id=sess-id-default",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--detach",
		"--",
		"interactive detach hello",
	}
	assertArgsEqual(t, resp.Args, want)
	assertNotHasArg(t, resp.Args, "--open")
	assertNotHasArg(t, resp.Args, "--new-terminal")
	if resp.StatusPollCalls < 1 {
		t.Fatal("expected wait-ready status poll")
	}
	if resp.LaunchCalls < 1 {
		t.Fatal("expected launch")
	}
}
```
