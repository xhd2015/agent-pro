# Scenario

**Feature**: React Router client navigation, NotFound, soft auth (Playwright)

```
# soft client routes — window marker survives; no full document reload
seed + token -> home click session-item -> /sessions/:id (chat-active)
session .back-link -> /
unknown path -> not-found -> not-found-home -> /
auth submit token -> home (soft gate, no location hard reload)
```

## Preconditions

- `playwright-debug` on PATH (else skip).
- Real frontend SPA preferred (React Router). Minimal dist shell alone cannot exercise client routes — implementer must ship structured SPA; leaves fail until then (TDD RED).
- Default explicit `--token test-token`.
- Label: `ui-automation` on every leaf ASSERT.

## Steps

1. Set `req.Mode = "ui"`.
2. Leaves seed, start web, compose Playwright script with nav marker.
3. `Run` executes playwright-debug.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
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
