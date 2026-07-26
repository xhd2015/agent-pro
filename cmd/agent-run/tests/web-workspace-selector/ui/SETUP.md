# Scenario

**Feature**: Playwright UI for `/workspace` selector and home integration

```
# SPA UI: open selector, draft survival, browse-only, Use commits, Cancel
token + home -> /workspace interactions
  -> only Use this folder commits selection
```

## Preconditions

- `playwright-debug` on PATH (else skip).
- Real frontend SPA preferred (React Router + WorkspacePage). Minimal dist
  shell alone cannot exercise routes — leaves fail until feature lands (TDD RED).
- Default explicit `--token test-token`.
- Label: `ui-automation` on every leaf ASSERT.
- Viewport 390×844.

## Steps

1. Set `req.Mode = "ui"`.
2. Leaves seed config/fixtures, start web, compose Playwright script.
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
