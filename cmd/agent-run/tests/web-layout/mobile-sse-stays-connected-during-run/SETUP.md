# Scenario

**Bug**: SSE stream must stay open for the full run — no idle abort when agent is silent for >2s

```
seed running session (user only, no live writer) -> session page -> SSE tails at offset -> 8s silence -> must not abort
```

## Preconditions

- Seeded `fake-opencode/sse-persist` with `status=running`, one user message in `events.jsonl`, no assistant reply.
- No live agent process appends events (forces >2s idle gap on open SSE).
- `playwright-debug` on PATH.
- Open API (`WebTokenMode=omit`).

## Steps

1. Seed running session via `seedRunningSessionAwaitingAssistant`; start `agent-run web` on free port.
2. Register `page.on('request')` + `requestfailed` counters **before** navigation.
3. Open session page; monitor network for 8s with no new events written server-side.
4. Assert exactly **1** `.../events/stream` request, **0** aborted/cancelled stream requests, session-detail GET **≤ 3**.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "sse-persistence"
	req.WebTokenMode = "omit"
	runner := "fake-opencode"
	sessionID := "sse-persist"

	if err := seedRunningSessionAwaitingAssistant(t, req.Home, runner, sessionID); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := initSSEPersistenceMonitor() + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'domcontentloaded' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + assertSSEStaysConnectedDuringRun()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```