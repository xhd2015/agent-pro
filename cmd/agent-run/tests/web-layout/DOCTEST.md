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

**Live agent harness** (Playwright leaves that POST or composer-trigger a real
run) builds `llm-mock-run-grok`, sets `LLM_MOCK_RUN_GROK_COMMAND` +
`AGENT_RUN_GROK_TTY_GROK_SESSION_ID`, and starts web with
`--agent-runner grok-tty --grok-home … --grok-tty-runner-binary …`.
Playwright scripts POST `{runner: "grok-tty", prompt: "…"}` — never
`fake-codex` for live sessions. Seeded-only leaves keep static `fake-codex`
fixture dirs unchanged.

**Session scroll / follow mode** (session detail page):

- `app-shell.session-page` uses `100dvh` with `overflow: hidden` — document must not scroll.
- Only `[data-testid="message-list"]` scrolls; `.top-bar`, `.session-header`,
  `.agent-running-card`, and `[data-testid="composer"]` stay fixed while the list moves.
- **Follow mode**: auto-scroll new/streaming content only while
  `distanceFromBottom = scrollHeight - scrollTop - clientHeight <= 80`.
- Scrolling up past threshold enters **detached** — freeze `scrollTop` on growth.
- Sending a message while detached does **not** re-enable follow or change scroll position.
- `[data-testid="jump-to-latest"]` appears when detached and content grows below;
  tap scrolls to bottom and resumes follow.

**Home scroll / follow mode** (home page `/`):

- `app-shell.home-page` uses `100dvh` with `overflow: hidden` — document must not scroll.
- Only `[data-testid="session-list"]` scrolls; `.top-bar-home` and
  `[data-testid="composer"]` stay fixed while the list moves.
- **Follow mode**: auto-scroll when poll refresh (`fetchSessions` every 3s) adds sessions
  only while `distanceFromBottom <= 80`.
- Scrolling up past threshold enters **detached** — freeze `scrollTop` on poll growth.
- Composer send while detached does **not** auto-scroll (navigation may follow separately).
- `[data-testid="jump-to-latest"]` appears when detached and new sessions appear below viewport.

**Layout selectors** (implemented by `frontend-agent-run/`):

| Selector | Role |
|----------|------|
| `[data-testid="app-shell"]` | Root flex column shell |
| `[data-testid="composer"]` | Message composer pinned to viewport bottom |
| `[data-testid="auth-page"]` | Token entry screen (401 / missing token) |
| `[data-testid="empty-state"]` | No sessions / empty chat welcome |
| `[data-testid="chat-active"]` | Active session with message list |
| `[data-testid="message-list"]` | Scrollable transcript region (only vertical scroll on session page) |
| `[data-testid="session-list"]` | Scrollable session list on home (only vertical scroll on home page) |
| `[data-testid="home-active"]` | Home main panel when session list is shown (fixed chrome wrapper) |
| `[data-testid="jump-to-latest"]` | Floating chip when detached with unseen content below |
| `.top-bar` | Session/home header chrome (fixed; no document scroll) |
| `.top-bar-home` | Home-specific top bar (title, runner, workspace) |
| `.session-header` | Session metadata block above transcript (fixed within panel) |
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
├── mobile-progress-card-compaction/      # duplicate think/tool rows compact; progress distinct from bubbles
├── mobile-progress-multi-tool-ordering/  # different tool_call_id interleave preserves start order
├── mobile-grok-tty-message-cards/        # grok-tty web_a1e886dbcebb3e2b-shaped message-card UX
├── mobile-running-status-card/        # session status=running → running card + duration
├── mobile-running-card-absent-when-idle/  # session status≠running → no running card
├── mobile-inline-assistant-loading/       # running + user only → inline loading bubble in list
├── mobile-streaming-assistant-bubble/     # live phased stream → assistant text length grows
├── mobile-follow-up-no-duplicate-user-messages/  # idle seed + composer follow-up → exactly 2 user bubbles
├── mobile-streaming-uses-sse-not-poll/    # live run → SSE tail; session-detail GET ≤3 in 8s window
├── mobile-sse-stays-connected-during-run/ # seeded running + no live writer → 1 SSE stream, 0 aborts in 8s idle gap
├── mobile-no-session-detail-poll-while-running/  # live fake-codex run; 15s passive watch → detail GET === 1, SSE === 1
├── mobile-home-runner-visible-long-workspace/  # long status.workspace on / → runner stays in viewport
├── mobile-session-messages-only-scroll/     # seeded overflow → document fixed; only message-list scrolls; chrome stable
├── mobile-session-auto-follow-at-bottom/    # live stream at bottom → scrollTop tracks growth
├── mobile-session-detach-on-scroll-up/      # scroll up during stream → scrollTop frozen
├── mobile-session-jump-to-latest/           # detached + growth → chip visible; tap restores follow
├── mobile-session-send-no-auto-scroll/      # detached + composer send → scrollTop unchanged
├── mobile-home-sessions-only-scroll/        # seeded ≥20 sessions → document fixed; only session-list scrolls; chrome stable
├── mobile-home-auto-follow-at-bottom/       # at bottom + poll adds session → scrollTop stays at bottom
├── mobile-home-detach-on-scroll-up/         # scroll up + poll adds session → scrollTop frozen
├── mobile-home-jump-to-latest/              # detached + poll growth → chip visible; tap restores follow
└── mobile-home-send-no-auto-scroll/         # detached + composer send → scrollTop unchanged before navigation
```

Parameter ranking (most → least significant):

1. **Route + screen** — home `/` vs session detail; empty vs auth vs chat vs running indicator
2. **Scroll / follow** — fixed chrome vs list-only scroll (message-list or session-list); following vs detached; jump chip
3. **Session status** — `running` (card visible) vs idle/finished (card absent)
4. **API token mode** — open API vs explicit Bearer
5. **Server cwd / workspace string length** — default vs deeply nested `WebWorkingDir` (home header)
6. **Viewport** — fixed 390×844 (iPhone-class)
7. **Layout invariants** — composer pinned bottom; `scrollWidth <= clientWidth`; runner picker within viewport width

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `mobile-empty` | No `--token`; no localStorage; empty-state (not auth page); composer pinned |
| 2 | `mobile-auth-page` | `--token test-token`; no localStorage; auth page; input pinned bottom |
| 3 | `mobile-chat-active` | `--token test-token` + seeded session; chat-active; composer pinned |
| 4 | `mobile-session-shows-workspace` | Session header workspace + distinct user message bubble |
| 5 | `mobile-message-roles-and-timestamps` | User vs assistant bubble styles differ; each message shows `message-timestamp` |
| 5a | `mobile-progress-card-compaction` | Duplicate think/tool rows compact; progress cards distinct from chat bubbles |
| 5b | `mobile-progress-multi-tool-ordering` | Interleaved tool_call_ids keep chronological card order on compaction |
| 5c | `mobile-grok-tty-message-cards` | Seeded `grok-tty/web_a1e886dbcebb3e2b` shape; role labels, bodies, progress separation |
| 6 | `mobile-running-status-card` | Seeded `running` session ~30s ago; `agent-running-card` + duration with digits |
| 7 | `mobile-home-runner-visible-long-workspace` | Web `cmd.Dir` = deep path; open `/` (no token); runner-picker in viewport |
| 8 | `mobile-running-card-absent-when-idle` | Seeded `idle` session; `agent-running-card` not in DOM (negative control) |
| 9 | `mobile-inline-assistant-loading` | Seeded `running` with user bubble only; inline `message-item-assistant-loading` visible |
| 10 | `mobile-streaming-assistant-bubble` | Live `grok-tty` create-session run; assistant bubble text length increases over poll window |
| 11 | `mobile-follow-up-no-duplicate-user-messages` | Seeded idle `grok-tty` session + 1 user event; composer follow-up; exactly 2 user bubbles after run |
| 12 | `mobile-streaming-uses-sse-not-poll` | Live `grok-tty` session; 8s network window: SSE used, session-detail GET ≤3 |
| 13 | `mobile-sse-stays-connected-during-run` | Seeded `running` session (no live writer); 8s idle gap: exactly 1 SSE stream, 0 aborts, detail GET ≤3 |
| 14 | `mobile-no-session-detail-poll-while-running` | Live `grok-tty` session; 15s passive watch: detail GET **=== 1**, SSE **=== 1**, 0 aborts (stricter than ≤3) |
| 15 | `mobile-session-messages-only-scroll` | Seeded `layout-scroll` (≥15 msgs); no document scroll; only `message-list` overflows; chrome Y stable after list scroll |
| 16 | `mobile-session-auto-follow-at-bottom` | Live `grok-tty` stream at bottom; `distanceFromBottom <= 80` on each assistant growth tick |
| 17 | `mobile-session-detach-on-scroll-up` | Scroll up ≥200px during stream; `scrollTop` frozen (±2px) while assistant text grows |
| 18 | `mobile-session-jump-to-latest` | Detached + streaming growth → `jump-to-latest` visible; tap → bottom + chip hidden |
| 19 | `mobile-session-send-no-auto-scroll` | Detached idle overflow session; composer send → `scrollTop` unchanged (±2px) |
| 20 | `mobile-home-sessions-only-scroll` | Seeded ≥20 home sessions; no document scroll; only `session-list` overflows; top-bar-home + composer Y stable after list scroll |
| 21 | `mobile-home-auto-follow-at-bottom` | At bottom on `/`; poll refresh adds 21st session; `distanceFromBottom <= 80` |
| 22 | `mobile-home-detach-on-scroll-up` | Scroll up ≥250px on `/`; poll adds session; `scrollTop` frozen (±2px) |
| 23 | `mobile-home-jump-to-latest` | Detached on `/` + poll growth → `jump-to-latest` visible; tap → bottom + chip hidden |
| 24 | `mobile-home-send-no-auto-scroll` | Detached on `/`; composer send → `scrollTop` unchanged (±2px) within 500ms |

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
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-session-messages-only-scroll --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-session-auto-follow-at-bottom --label 'chromium && slow'
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-session-detach-on-scroll-up --label 'chromium && slow'
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-session-jump-to-latest --label 'chromium && slow'
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-session-send-no-auto-scroll --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-home-sessions-only-scroll --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-home-auto-follow-at-bottom --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-home-detach-on-scroll-up --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-home-jump-to-latest --label chromium
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-home-send-no-auto-scroll --label chromium
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
	LLMMockRunGrok      string
	GrokHome            string
	GrokTTYRunnerBinary string
	Port       int
	Token         string
	WebTokenMode  string // "omit" | "explicit" | "auto" (default explicit)
	BaseURL       string
	Env        []string
	Layout     string // empty | auth | chat-active | running-card | running-absent | home-long-workspace | follow-up-dedupe | sse-transport | sse-persistence | no-detail-poll | streaming-bubble | inline-loading | workspace-session | session-messages-only-scroll | session-auto-follow-at-bottom | session-detach-on-scroll-up | session-jump-to-latest | session-send-no-auto-scroll | home-sessions-only-scroll | home-auto-follow-at-bottom | home-detach-on-scroll-up | home-jump-to-latest | home-send-no-auto-scroll
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

	cmd := exec.CommandContext(ctx, "playwright-debug", "-e", script)
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