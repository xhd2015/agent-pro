# Web Layout Tests

Playwright mobile layout tests for `agent-run web` (viewport 390×844). Each leaf
starts the embedded SPA against a temp `AGENT_RUN_HOME`, then asserts layout
invariants via `playwright-debug`.

# DSN (Domain Specific Notion)

**agent-run web** serves the mobile-first SPA from `frontend-agent-run/` on
`127.0.0.1:<port>`. Static assets are always unauthenticated. API auth follows
the same three `--token` modes as CLI tests:

| Mode | CLI | SPA behavior |
|------|-----|--------------|
| Open | (no `--token`) | Health succeeds without Bearer; app skips auth page |
| Explicit | `--token <value>` | 401 without matching Bearer; auth page when storage empty |
| Auto | `--token auto` | Generated token on stderr; Bearer required |

The **test harness** builds `cmd/agent-run`, allocates a free port, starts
`agent-run web` in the background per leaf token mode, waits for health, then
runs a **playwright-debug** script from `req.PlaywrightScript`.

**Layout selectors** (implemented by `frontend-agent-run/`):

| Selector | Role |
|----------|------|
| `[data-testid="app-shell"]` | Root flex column shell |
| `[data-testid="composer"]` | Message composer pinned to viewport bottom |
| `[data-testid="auth-page"]` | Token entry screen (401 / missing token) |
| `[data-testid="empty-state"]` | No sessions / empty chat welcome |
| `[data-testid="chat-active"]` | Active session with message list |
| `[data-testid="message-list"]` | Scrollable transcript region |
| `[data-testid="workspace"]` | Session or landing workspace path |
| `[data-testid="message-item-user"]` | User-authored timeline bubble |
| `[data-testid="message-item-assistant"]` | Assistant timeline bubble |
| `[data-testid="message-timestamp"]` | Per-message local time label |
| `[data-testid="agent-running-card"]` | Prominent “agent working” block when session status is `running` |
| `[data-testid="agent-running-duration"]` | Live elapsed-time text inside the running card |
| `[data-testid="message-item-assistant-loading"]` | Inline pulsing assistant placeholder below last user message while `running` |
| `[data-testid="runner-picker"]` | Home/session runner `<label>` wrapping the select |
| `[data-testid="runner-select"]` | Runner `<select>` control |

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web-layout/
├── DOCTEST.md
├── SETUP.md                           # buildAgentRun, findFreePort, startWebBackground, requirePlaywright
├── mobile-empty/                      # open API (no --token); no auth page → empty state + composer
├── mobile-auth-page/                  # explicit --token; no localStorage → auth page
├── mobile-chat-active/                # explicit --token + seeded session → chat-active
├── mobile-session-shows-workspace/    # seeded workspace + user/assistant roles in transcript
├── mobile-message-roles-and-timestamps/  # distinct bubble styles + timestamp testids
├── mobile-running-status-card/        # session status=running → running card + duration
├── mobile-running-card-absent-when-idle/  # session status≠running → no running card
├── mobile-inline-assistant-loading/       # running + user only → inline loading bubble in list
├── mobile-streaming-assistant-bubble/     # live phased stream → assistant text length grows
├── mobile-follow-up-no-duplicate-user-messages/  # idle seed + composer follow-up → exactly 2 user bubbles
├── mobile-streaming-uses-sse-not-poll/    # live run → SSE tail; session-detail GET ≤3 in 8s window
├── mobile-sse-stays-connected-during-run/ # seeded running + no live writer → 1 SSE stream, 0 aborts in 8s idle gap
├── mobile-no-session-detail-poll-while-running/  # live fake-codex run; 15s passive watch → detail GET === 1, SSE === 1
└── mobile-home-runner-visible-long-workspace/  # long status.workspace on / → runner stays in viewport
```

Parameter ranking (most → least significant):

1. **Route + screen** — home `/` vs session detail; empty vs auth vs chat vs running indicator
2. **Session status** — `running` (card visible) vs idle/finished (card absent)
3. **API token mode** — open API vs explicit Bearer
4. **Server cwd / workspace string length** — default vs deeply nested `WebWorkingDir` (home header)
5. **Viewport** — fixed 390×844 (iPhone-class)
6. **Layout invariants** — composer pinned bottom; `scrollWidth <= clientWidth`; runner picker within viewport width

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `mobile-empty` | No `--token`; no localStorage; empty-state (not auth page); composer pinned |
| 2 | `mobile-auth-page` | `--token test-token`; no localStorage; auth page; input pinned bottom |
| 3 | `mobile-chat-active` | `--token test-token` + seeded session; chat-active; composer pinned |
| 4 | `mobile-session-shows-workspace` | Session header workspace + distinct user message bubble |
| 5 | `mobile-message-roles-and-timestamps` | User vs assistant bubble styles differ; each message shows `message-timestamp` |
| 6 | `mobile-running-status-card` | Seeded `running` session ~30s ago; `agent-running-card` + duration with digits |
| 7 | `mobile-home-runner-visible-long-workspace` | Web `cmd.Dir` = deep path; open `/` (no token); runner-picker in viewport |
| 8 | `mobile-running-card-absent-when-idle` | Seeded `idle` session; `agent-running-card` not in DOM (negative control) |
| 9 | `mobile-inline-assistant-loading` | Seeded `running` with user bubble only; inline `message-item-assistant-loading` visible |
| 10 | `mobile-streaming-assistant-bubble` | Live create-session run; assistant bubble text length increases over poll window |
| 11 | `mobile-follow-up-no-duplicate-user-messages` | Seeded idle session + 1 user event; composer follow-up; exactly 2 user bubbles after run |
| 12 | `mobile-streaming-uses-sse-not-poll` | Live `fake-codex` session; 8s network window: SSE used, session-detail GET ≤3 |
| 13 | `mobile-sse-stays-connected-during-run` | Seeded `running` session (no live writer); 8s idle gap: exactly 1 SSE stream, 0 aborts, detail GET ≤3 |
| 14 | `mobile-no-session-detail-poll-while-running` | Live `fake-codex` session; 15s passive watch: detail GET **=== 1**, SSE **=== 1**, 0 aborts (stricter than ≤3) |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/web-layout
doctest test -v ./cmd/agent-run/tests/web-layout
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-empty
doctest test -v ./cmd/agent-run/tests/web-layout --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-running-status-card --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-home-runner-visible-long-workspace --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-follow-up-no-duplicate-user-messages --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-streaming-uses-sse-not-poll --label 'chromium && slow'
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-sse-stays-connected-during-run --label 'chromium && slow'
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-no-session-detail-poll-while-running --label 'chromium && slow'
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Request struct {
	TempDir    string
	Home       string
	RepoRoot   string
	AgentRun   string
	Port       int
	Token         string
	WebTokenMode  string // "omit" | "explicit" | "auto" (default explicit)
	BaseURL       string
	Env        []string
	Layout     string // empty | auth | chat-active | running-card | running-absent | home-long-workspace | follow-up-dedupe | sse-transport | sse-persistence | no-detail-poll | streaming-bubble | inline-loading | workspace-session
	PlaywrightScript string
	WebWorkingDir    string // optional process cwd for agent-run web (status.workspace)

	webCmd *exec.Cmd
}

type Response struct {
	PlaywrightStdout string
	PlaywrightStderr string
	PlaywrightExit   int
	Err              error
}

func runPlaywrightScript(t *testing.T, script string) (string, string, int, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "playwright-debug", "run", script)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout := outBuf.String()
	stderr := errBuf.String()
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return stdout, stderr, exitErr.ExitCode(), nil
		}
		if ctx.Err() != nil {
			return stdout, stderr, -1, ctx.Err()
		}
		return stdout, stderr, -1, runErr
	}
	return stdout, stderr, 0, nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	requirePlaywright(t)

	if req.webCmd == nil && req.Port > 0 {
		if err := startWebBackground(t, req); err != nil {
			return nil, err
		}
	}
	if req.BaseURL == "" && req.Port > 0 {
		req.BaseURL = "http://127.0.0.1:" + strconv.Itoa(req.Port)
	}

	if strings.TrimSpace(req.PlaywrightScript) == "" {
		return nil, fmt.Errorf("PlaywrightScript is empty — leaf Setup must set it")
	}

	stdout, stderr, exitCode, err := runPlaywrightScript(t, req.PlaywrightScript)
	resp := &Response{
		PlaywrightStdout: stdout,
		PlaywrightStderr: stderr,
		PlaywrightExit:   exitCode,
		Err:              err,
	}
	return resp, err
}
```