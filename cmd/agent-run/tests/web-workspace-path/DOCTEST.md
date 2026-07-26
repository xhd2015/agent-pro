# Web Workspace Path Tap-to-Expand Tests

Playwright mobile (390×844) tests for option A: compact workspace label with
tap-to-expand full path and copy-to-clipboard, on both home top bar and session
header. Reuses web-layout harness patterns (`startWebBackground`, deep
`WebWorkingDir`, `playwright-debug`).

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web** — serves the mobile-first SPA (`frontend-agent-run`) on
  `127.0.0.1:<port>`; `GET /api/agent-run/status` exposes `workspace` (server
  process cwd); session meta stores per-session `workspace`.
- **WorkspacePath UI** — shared control in home top bar and session header:
  root `[data-testid="workspace"]`, visible text `[data-testid="workspace-label"]`,
  expand/collapse `[data-testid="workspace-toggle"]` (may be the label), copy
  `[data-testid="workspace-copy"]`. Local React state `expanded` (not persisted).
- **shortWorkspaceLabel** — collapsed display: last ≤2 path segments; deep paths
  become `…/last/two`. Full absolute path remains the source of truth for
  expand/copy/`title`.
- **Runner picker** — `[data-testid="runner-picker"]` / `runner-select` on home;
  must stay within the 390px viewport while the path is **collapsed**.
- **Test harness** — builds `cmd/agent-run`, temp `AGENT_RUN_HOME`, free port,
  optional deep `WebWorkingDir` for status.workspace, seeds flat
  `sessions/<id>/` for session leaves, runs `playwright-debug` scripts.

**Behaviors**

```
# Home default (collapsed): shortened label; runner stays in viewport
Browser -> GET / -> status.workspace (long) -> WorkspacePath collapsed
  -> label shows …/last/two; runner-picker within 390px

# Tap expand: full path, multi-line OK; no horizontal document scroll
User -> tap workspace-toggle -> aria-expanded=true
  -> workspace-label text === full absolute path
  -> documentElement.scrollWidth <= clientWidth

# Second tap: collapse back to compact label
User -> tap workspace-toggle -> aria-expanded=false -> short label again

# Copy: clipboard gets full path
User -> (expanded) tap workspace-copy -> navigator.clipboard === full path

# Session page: same WorkspacePath component; meta.workspace is source
Browser -> GET /sessions/<id> -> session.workspace long
  -> collapsed short; tap expand -> full path visible
```

**Selectors**

| Selector | Role |
|----------|------|
| `[data-testid="workspace"]` | Root region (existing id; keep) |
| `[data-testid="workspace-label"]` | Visible path text (compact or full) |
| `[data-testid="workspace-toggle"]` | Expand/collapse control (`aria-expanded`) |
| `[data-testid="workspace-copy"]` | Copy full path (at least when expanded) |
| `[data-testid="runner-picker"]` | Home runner control (viewport regression) |
| `[data-testid="runner-select"]` | Runner `<select>` |
| `[data-testid="empty-state"]` | Home empty panel (no sessions) |
| `[data-testid="chat-active"]` | Session detail main panel |

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web-workspace-path/
├── DOCTEST.md
├── SETUP.md                              # build agent-run, deep/short workspace helpers, playwright
├── home/                                 # surface = home page /
│   ├── SETUP.md                          # open API (omit token); home scenario defaults
│   ├── long-path/                        # WebWorkingDir = deep nested cwd
│   │   ├── SETUP.md
│   │   ├── collapsed-default/            # short …/ label; runner-picker in viewport
│   │   ├── tap-expand/                   # tap toggle → full path; no h-scroll  [expect RED]
│   │   ├── tap-collapse/                 # expand then collapse → short again   [expect RED]
│   │   └── copy-full-path/               # expand + copy → clipboard full path  [expect RED]
│   └── short-path/                       # cwd ≤2 meaningful segments
│       ├── SETUP.md
│       └── collapsed-readable/           # label readable; short === full semantics
└── session-detail/                       # surface = session detail
    ├── SETUP.md                          # explicit token; flat session seed helper
    └── long-path/
        ├── SETUP.md                      # meta.workspace = deep path
        └── tap-expand/                   # collapsed short + expand full         [expect RED]
```

Parameter ranking (most → least significant):

1. **Surface** — home `/` (status.workspace + runner chrome) vs session detail (meta.workspace).
2. **Path length** — long (needs `…/` collapse) vs short (≤2 segments; expand optional).
3. **Interaction** — default collapsed vs tap expand vs tap collapse vs copy clipboard.
4. **Viewport** — fixed 390×844; no horizontal document scroll (especially when expanded).
5. **Token mode** — home open API; session explicit Bearer (SPA auth gate).

## Test Index

| # | Leaf | Description | Expect RED today? |
|---|------|-------------|-------------------|
| 1 | `home/long-path/collapsed-default` | Long status.workspace on `/`; shortened `…/` label; runner-picker within viewport | Partial green (short label + runner already true); new child testids optional |
| 2 | `home/long-path/tap-expand` | Tap `[data-testid="workspace-toggle"]` → full path in label; `aria-expanded=true`; no h-scroll | **Yes** |
| 3 | `home/long-path/tap-collapse` | Expand then second tap → short label; `aria-expanded=false`; runner still in viewport | **Yes** |
| 4 | `home/long-path/copy-full-path` | Expand + click `[data-testid="workspace-copy"]`; clipboard === full path | **Yes** |
| 5 | `home/short-path/collapsed-readable` | Short cwd (≤2 segments); label readable and equals full path form | No (label already shows short path) |
| 6 | `session-detail/long-path/tap-expand` | Seeded long `meta.workspace`; collapsed short then expand shows full path | **Yes** |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/web-workspace-path                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/web-workspace-path
doctest test --label-all ./cmd/agent-run/tests/web-workspace-path

doctest vet ./cmd/agent-run/tests/web-workspace-path
doctest test ./cmd/agent-run/tests/web-workspace-path
doctest test -v ./cmd/agent-run/tests/web-workspace-path
doctest test --label ui-automation ./cmd/agent-run/tests/web-workspace-path
doctest test -v ./cmd/agent-run/tests/web-workspace-path/home/long-path/tap-expand --label ui-automation
doctest test -v ./cmd/agent-run/tests/web-workspace-path/home/long-path/copy-full-path --label ui-automation
doctest test -v ./cmd/agent-run/tests/web-workspace-path/session-detail/long-path/tap-expand --label ui-automation
```

Regression companion (existing tree, not re-owned here):

```sh
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-home-runner-visible-long-workspace --label chromium
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	TempDir          string
	Home             string
	RepoRoot         string
	AgentRun         string
	Port             int
	Token            string
	WebTokenMode     string // "omit" | "explicit" | "auto" (default explicit)
	BaseURL          string
	Env              []string
	Scenario         string // leaf scenario id for Assert checks
	PlaywrightScript string
	WebWorkingDir    string // optional process cwd for agent-run web (status.workspace)
	WorkspacePath    string // full absolute path under test (home cwd or session meta)
	SessionID        string

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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
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
