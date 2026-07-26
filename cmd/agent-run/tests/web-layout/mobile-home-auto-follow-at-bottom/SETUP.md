# Scenario

**Feature**: home follow mode — auto-scroll while user is at top (newest-first) during poll refresh

```
seed 20 home sessions → / at top → append 21st session dir → poll refresh → session-list stays at top
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded 20 home sessions; `session-list` overflows mobile viewport.
- Open API (`WebTokenMode=omit`).
- 21st session dir appended after page load; home poll interval ~3s.
- Home list is **newest-first**; follow/jump target is the **top**.

## Steps

1. Seed 20 home sessions under `fake-codex` (flat `sessions/<id>/`).
2. Start web with open API; open `/`.
3. Ensure `session-list` at top (`distanceFromTop <= 80`).
4. Schedule append of `home-sess-021` after page settles.
5. Wait 4s for poll refresh; assert `session-list` remains at top.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-auto-follow-at-bottom"
	req.WebTokenMode = "omit"
	runner := "fake-codex"

	if err := seedManyHomeSessions(t, req.Home, runner, 20); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	scheduleAppendHomeSessionDir(t, req, runner, 21, 1500*time.Millisecond)

	body := openHomePage(req.BaseURL) + waitForSessionListOverflow() +
		scrollSessionListToTop() + assertSessionListAtTop() +
		waitForHomePollRefresh() + assertSessionListAtTop()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
