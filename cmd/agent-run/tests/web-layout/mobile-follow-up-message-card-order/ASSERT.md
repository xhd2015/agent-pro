---
label: chromium, slow
explanation: Seeded idle + one live grok-tty follow-up; order poll ≤60s + reload
---

## Expected

- Playwright exit code **0**.
- Viewport 390×844; no horizontal document scroll.
- Seeded session shows first user `run ls` and at least one assistant before follow-up.
- After follow-up send, once both user prompts are visible in the live timeline:
  - First user bubble contains `run ls`, second contains `what did I say` (chronological).
  - Each of those two prompts appears in **exactly one** user bubble.
  - No non-empty `[data-testid="message-item-assistant"]` appears **before** the first user bubble (anti-regression for strip-all-users merge that jumped users under assistants).
- After full page reload of the session URL, the same order invariants hold.

## Side Effects

- Seeded session under flat `AGENT_RUN_HOME/sessions/follow-up-card-order/`.
- One live follow-up agent run via `grok-tty` mock harness (`llm-mock-run-grok`).

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
	if req.Layout != "follow-up-message-card-order" {
		t.Fatalf("expected layout follow-up-message-card-order, got %q", req.Layout)
	}
}
```
