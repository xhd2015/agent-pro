# Scenario

**Feature**: home follow mode — detach when user scrolls down (away from newest) during poll refresh

```
seed 20 home sessions → scroll down to detach → append 21st session → poll → scrollTop frozen
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded 20 home sessions with overflow list.
- Open API.
- 21st session appended after detach; poll refresh must not move `scrollTop`.
- Home list is **newest-first**; detach = scroll **down** from top.

## Steps

1. Seed 20 home sessions under `fake-codex` (flat layout).
2. Open `/`; wait for `session-list` overflow.
3. Scroll `session-list` down ≥250px from top; record `scrollTop`.
4. Schedule append of `home-sess-021`.
5. Wait 4s for poll; assert `scrollTop` stable (±2px).

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-detach-on-scroll-up"
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
		scrollSessionListDownFromTop(250) + assertSessionListDetached() +
		recordSessionListScrollTop("FrozenScrollTop") + waitForHomePollRefresh() +
		assertSessionListScrollTopEqualsVar("FrozenScrollTop")

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
