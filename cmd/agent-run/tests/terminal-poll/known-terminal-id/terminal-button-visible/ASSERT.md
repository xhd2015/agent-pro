---
label: ui-automation
explanation: AX tree poll for Terminal button on finished session with known mapping
---

## Expected

- Playwright exit code **0**.
- Finished `grok-tty` chat page shows an **enabled** Terminal button.
- Status pill reads **finished**.

## Side Effects

- None beyond seeded fixtures and browser navigation.

## Exit Code

- Playwright process exits 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Scenario != "terminal-button-visible" {
		t.Fatalf("expected scenario terminal-button-visible, got %q", req.Scenario)
	}
}
```