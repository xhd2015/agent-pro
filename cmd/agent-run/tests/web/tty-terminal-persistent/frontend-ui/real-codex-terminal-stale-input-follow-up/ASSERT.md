---
label: codex, ui-automation, slow, flaky
explanation: Requires the real codex CLI, network/model availability, and Playwright browser automation.
---

## Expected

- Playwright exits 0.
- The test uses the real `codex` binary on `PATH`, not a fake fixture. The TTY
  command override points at that binary only to pass Codex CLI flags needed for
  non-interactive test startup.
- A terminal draft typed but not submitted before closing the modal is cleared
  before the chat follow-up is sent.
- The real Codex terminal does not exit or print only a resume footer before
  producing the first assistant answer.
- The chat assistant response after the follow-up does not contain raw terminal
  scrollback, Codex banner text, terminal session JSON, prompt echo, or terminal
  exit/status text.

## Exit Code

- Test process exits non-zero while real Codex exits before the first answer,
  stale terminal input is submitted with the follow-up, or raw terminal output
  is persisted/rendered as the chat assistant response.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s", resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}
```
