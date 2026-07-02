---
label: ui-automation
explanation: Requires playwright-debug and browser automation.
---

## Expected

- Playwright exits 0.
- A running `codex-tty` chat shows the terminal button before the assistant
  response finishes.
- The button is enabled while the server-side tty is already available.

## Exit Code

- Test process exits non-zero until the frontend derives terminal affordance
  from terminal availability immediately during active tty runs.

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
