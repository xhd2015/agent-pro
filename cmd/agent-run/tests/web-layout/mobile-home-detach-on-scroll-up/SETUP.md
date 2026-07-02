# Scenario

**Feature**: home follow mode — detach when user scrolls up during poll refresh

```
seed 20 home sessions → scroll up to detach → append 21st session → poll → scrollTop frozen
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded 20 home sessions with overflow list.
- Open API.
- 21st session appended after detach; poll refresh must not move `scrollTop`.

## Steps

1. Seed 20 home sessions under `fake-codex`.
2. Open `/`; wait for `session-list` overflow.
3. Scroll `session-list` up ≥250px from bottom; record `scrollTop`.
4. Schedule append of `home-sess-021`.
5. Wait 4s for poll; assert `scrollTop` stable (±2px).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
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
		scrollSessionListUpFromBottom(250) + assertSessionListDetached() +
		recordSessionListScrollTop("FrozenScrollTop") + waitForHomePollRefresh() +
		assertSessionListScrollTopEqualsVar("FrozenScrollTop")

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```