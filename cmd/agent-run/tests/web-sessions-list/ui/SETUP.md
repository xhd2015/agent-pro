# Scenario

**Feature**: Playwright UI for home sessions list (newest-first, search, load more)

```
# SPA home: seed metas -> / -> assert list order / search / load-more
token + home sessions
  -> session-list newest first
  -> session-search filters
  -> session-load-more button inside list; click appends (no infinite scroll)
```

## Preconditions

- `playwright-debug` on PATH (else skip).
- Real frontend SPA preferred. Minimal dist shell alone cannot exercise
  HomePage — leaves fail until feature lands (TDD RED).
- Default explicit `--token test-token`.
- Label: `ui-automation` on every leaf ASSERT.
- Viewport 390×844.

## Steps

1. Set `req.Mode = "ui"`.
2. Leaves seed sessions, start web, compose Playwright script.
3. `Run` executes playwright-debug.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Mode = "ui"
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebTokenMode == "explicit" && req.Token == "" {
		req.Token = "test-token"
	}
	return nil
}
```
