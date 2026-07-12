## Expected

Exact argv includes open profile plus `-e` env pairs **after** `--open` and
**before** `--` / prompt:

```text
run
--session-id=sess-env-1
--agent-runner=grok-tty
--auto-send-or-resume
--new-terminal
--open
-e
SLACK_MSG_SESSION_ID=sess-env-1
-e
SLACK_MSG_CONFIG=/tmp/slack-config.json
--
with env
```

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A (package call)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	want := []string{
		"run",
		"--session-id=sess-env-1",
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
		"--open",
		"-e",
		"SLACK_MSG_SESSION_ID=sess-env-1",
		"-e",
		"SLACK_MSG_CONFIG=/tmp/slack-config.json",
		"--",
		"with env",
	}
	assertArgsEqual(t, resp.Args, want)
}
```
