# Scenario

**Feature**: idle home does not periodically re-fetch runners or status after bootstrap

```
open / → bootstrap may GET sessions + runners + status once
  → idle ~7s → runners GET Δ=0, status GET Δ=0
  (sessions poll may be sparse; idle cadence ~15s)
```

## Preconditions

- `playwright-debug` on PATH.
- Open API; seed a few sessions so home shows list.
- Label: `ui-automation, slow` (~7s idle wait).

## Steps

1. Seed 5 home sessions; start web open API.
2. Open `/` with request counters; wait for home; snapshot counters after bootstrap settle.
3. Idle 7s; assert runnersΔ=0 and statusΔ=0.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-idle-no-meta-poll"
	req.WebTokenMode = "omit"
	runner := "fake-codex"

	if err := seedManyHomeSessions(t, req.Home, runner, 5); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := `
const apiHits = { sessions: 0, runners: 0, status: 0 };
page.on('request', (req) => {
  if (req.method() !== 'GET') return;
  const u = req.url();
  if (!u.includes('/api/agent-run/')) return;
  if (u.includes('/api/agent-run/sessions')) apiHits.sessions += 1;
  else if (u.includes('/api/agent-run/runners')) apiHits.runners += 1;
  else if (u.includes('/api/agent-run/status')) apiHits.status += 1;
});
` + openHomePage(req.BaseURL) + `
await page.waitForTimeout(500);
const afterBootstrap = {
  sessions: apiHits.sessions,
  runners: apiHits.runners,
  status: apiHits.status,
};
// Bootstrap may hit runners/status once each; reject wasteful multi-hit already.
if (afterBootstrap.runners > 2 || afterBootstrap.status > 2) {
  throw new Error('unexpected bootstrap meta fan-out: ' + JSON.stringify(afterBootstrap));
}
await page.waitForTimeout(7000);
const afterIdle = {
  sessions: apiHits.sessions,
  runners: apiHits.runners,
  status: apiHits.status,
  sessionsDelta: apiHits.sessions - afterBootstrap.sessions,
  runnersDelta: apiHits.runners - afterBootstrap.runners,
  statusDelta: apiHits.status - afterBootstrap.status,
};
if (afterIdle.runnersDelta !== 0 || afterIdle.statusDelta !== 0) {
  throw new Error(
    'home idle poll must not re-fetch runners/status: bootstrap=' +
      JSON.stringify(afterBootstrap) + ' idle=' + JSON.stringify(afterIdle),
  );
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
