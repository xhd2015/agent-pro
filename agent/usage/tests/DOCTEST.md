# agent/usage facade tests

Doc-style tests for `github.com/xhd2015/agent-pro/agent/usage`, the unified
in-process usage fetch facade over Grok and Codex TTY providers.

# DSN (Domain Specific Notion)

**Caller** (ai-critic menubar services or other libraries) invokes
`usage.Fetch(ctx, providerID)` and receives a normalized **Snapshot**.

**usage.Fetch** dispatches by `ProviderID`:

- `Grok` → `agent/grok/tty.FetchUsage` (direct PTY; honors `GROK_SHOW_USAGE_COMMAND`)
- `Codex` → `agent/codex/tty.FetchStatus` (in-process ttywatch session; honors
  `CODEX_SHOW_STATUS_COMMAND` and `TTY_WATCH_HOME`)

**Fake TUI hooks** replace production agent binaries for deterministic doctests,
matching the existing CLI test fixtures.

```
caller -> usage.Fetch(ctx, Grok|Codex)
  -> provider tty helper -> fake TUI (env command hook) -> parse -> Snapshot
doctest <- Snapshot fields (provider, usage %, reset, credits)
```

Tests call `usage.Fetch` directly (no CLI subprocess).

## Version

0.0.2

## Decision Tree

```
agent/usage/tests/
├── DOCTEST.md
├── SETUP.md                           # env helpers, fake TUI scripts, isolated TTY_WATCH_HOME
├── fetch-grok-mock/                   # GROK_SHOW_USAGE_COMMAND fake → weekly limit + reset
└── fetch-codex-mock/                  # CODEX_SHOW_STATUS_COMMAND fake → monthly % + credits + reset
```

Parameter ranking (most → least significant):

1. **Provider** — Grok vs Codex (different tty backends and env hooks)
2. **Runner backend** — fake TUI via env command hook (deterministic fixtures)
3. **Snapshot fields** — provider-specific parsed values in normalized struct

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `fetch-grok-mock` | `Fetch(ctx, Grok)` with fake hook returns weekly 1% + July 9 reset (RED) |
| 2 | `fetch-codex-mock` | `Fetch(ctx, Codex)` with fake hook returns 58% usage, credits 6519/11250, reset (RED) |

## How to Run

```sh
doctest vet ./agent/usage/tests
doctest test ./agent/usage/tests/...
doctest test -v ./agent/usage/tests/fetch-grok-mock
doctest test -v ./agent/usage/tests/fetch-codex-mock
```

```go
import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
)

type Request struct {
	Provider          usage.ProviderID
	TTYWatchHome      string
	ShowUsageCommand  string // GROK_SHOW_USAGE_COMMAND
	ShowStatusCommand string // CODEX_SHOW_STATUS_COMMAND
	SessionID         string // CODEX_SHOW_STATUS_SESSION_ID; default codex-status-usage
	TimeoutSeconds    string // provider timeout env override; empty = default
}

type Response struct {
	Snapshot *usage.Snapshot
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Provider == "" {
		t.Fatal("req.Provider must be set by leaf Setup")
	}

	timeout := 75 * time.Second
	if req.TimeoutSeconds != "" {
		if sec, err := time.ParseDuration(req.TimeoutSeconds + "s"); err == nil {
			timeout = sec + 20*time.Second
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if req.TTYWatchHome != "" {
		t.Setenv("TTY_WATCH_HOME", req.TTYWatchHome)
	}
	if req.ShowUsageCommand != "" {
		t.Setenv("GROK_SHOW_USAGE_COMMAND", req.ShowUsageCommand)
	}
	if req.ShowStatusCommand != "" {
		t.Setenv("CODEX_SHOW_STATUS_COMMAND", req.ShowStatusCommand)
	}
	if req.TimeoutSeconds != "" {
		switch req.Provider {
		case usage.Grok:
			t.Setenv("GROK_SHOW_USAGE_TIMEOUT", req.TimeoutSeconds)
		case usage.Codex:
			t.Setenv("CODEX_SHOW_STATUS_TIMEOUT", req.TimeoutSeconds)
		}
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		sid = "codex-status-usage"
	}
	t.Setenv("CODEX_SHOW_STATUS_SESSION_ID", sid)

	snap, err := usage.Fetch(ctx, req.Provider)
	if err != nil {
		return nil, err
	}
	return &Response{Snapshot: snap}, nil
}
```