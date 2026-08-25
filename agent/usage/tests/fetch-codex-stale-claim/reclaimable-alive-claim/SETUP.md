# Scenario

**Feature**: leftover alive provisional claim for `codex-status-usage` must not block usage Fetch

```
write registry/.codex-status-usage.claim  # PID = test process (alive)
usage.Fetch(ctx, Codex) + fake TUI
doctest <- Snapshot (same fields as fetch-codex-mock)
```

## Preconditions

- Inherits group Codex + fake TUI + SessionID.
- `PlantAliveSessionClaim=true` so Run plants the claim before Fetch.

## Steps

1. Enable `PlantAliveSessionClaim`.
2. Run Fetch; Assert expects success snapshot.

## Context

- Was RED when FetchStatus reserved without reclaim; now green via
  `ReserveCustomSessionIDReclaiming` (same as headless).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.PlantAliveSessionClaim = true
	return nil
}
```
