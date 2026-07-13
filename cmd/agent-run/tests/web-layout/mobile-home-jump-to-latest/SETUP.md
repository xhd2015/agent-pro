# Scenario

**Feature**: home jump-to-latest chip restores follow at top after detach

```
seed 20 home sessions → scroll down to detach → append session → poll → chip visible → tap → top
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded 20 home sessions with overflow list.
- Open API.
- New session above viewport while detached shows `[data-testid="jump-to-latest"]`.
- Home list is **newest-first**; jump-to-latest scrolls to **top**.

## Steps

1. Seed 20 home sessions under `fake-codex` (flat layout).
2. Open `/`; wait for `session-list` overflow.
3. Scroll `session-list` down ≥250px from top to detach.
4. Schedule append of `home-sess-021`; wait for poll refresh.
5. Wait for `[data-testid="jump-to-latest"]`; tap; assert top + chip hidden.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-jump-to-latest"
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
		waitForHomePollRefresh() + assertHomeJumpToLatestChipFlow()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
