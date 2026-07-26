# Web Workspace Path Selector Tests

Doctests for the SPA path selector (`/workspace`) and backend
`selected_workspace` / MRU persistence. Covers HTTP APIs and Playwright UI,
with the critical product rule: **only “Use this folder” commits** (PUT);
quick chips and Recent only change browse path.

Reuses harness patterns from `web-layout`, `web-workspace-path`, and
`web-spa-structured-app` (`startWebBackground`, free port, `playwright-debug`,
flat session layout).

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web** — binds `127.0.0.1:<port>`; serves SPA + `/api/agent-run/*`.
  Isolated per-leaf via `AGENT_RUN_HOME` and optional `WebWorkingDir` (process cwd).
- **config.json** (`AGENT_RUN_HOME/config.json`) — persists
  `selected_workspace` (active path) and `recent_workspaces` (backend MRU,
  max ~12, newest first, move-to-front on Use). No frontend localStorage MRU.
- **Status API** — `GET /api/agent-run/status` exposes resolved `workspace`,
  plus `process_cwd`, `home` (OS user home for Quick Home), and
  `recent_workspaces`.
- **Workspace API** — `PUT /api/agent-run/workspace` with `{"path":"..."}`
  validates absolute existing **directory**, sets selected, updates MRU,
  returns status-shaped JSON.
- **FS list API** — `GET /api/agent-run/fs/list?path=` lists dir + file entries
  with `type: "dir"|"file"` and parent path. Includes **dot dirs and dot files**;
  **dirs sorted before files**, case-insensitive name order within each group.
- **Sessions API** — `POST /api/agent-run/sessions` stamps
  `meta.workspace` from resolved selected workspace (fallback process cwd).
- **HomePage** — workspace control opens selector (`/workspace`); composer
  draft is **app-level** so it survives unmount during selector round-trip.
- **WorkspacePage** (`/workspace`) — Cancel | browse path | Quick chips |
  Recent rows | Parent | dir list (files **hidden by default**) |
  **Show files** toggle (`workspace-show-files`, `aria-expanded`) placed
  **after the last directory entry** and **before any file rows** when expanded |
  **Use this folder** (sole commit). Local `showFiles` resets on path change.
- **Test harness** — builds `agent-run`, temp home, free port, multi-step HTTP
  and/or `playwright-debug` scripts.

**Behaviors**

```
# Resolve workspace when none selected
GET /api/agent-run/status -> workspace ≈ process cwd; recent empty/absent

# Commit selection (API)
PUT /api/agent-run/workspace {"path":"/abs/dir"}
  -> 200; selected_workspace set; path first in recent_workspaces
  -> config.json persisted; next GET /status still selected

# Reject bad paths
PUT file path | missing path -> 400; config unchanged

# MRU
PUT existing recent path -> move-to-front; length <= 12

# New session uses selected
PUT selected -> POST /sessions -> response/meta.workspace == selected

# List filesystem (optimize)
GET /fs/list?path=<dir>
  -> entries: dirs then files; include .git/.env; case-insensitive within group

# UI: open selector (browse only until Use)
Home -> tap workspace open control -> URL /workspace; selector visible

# UI: draft safety
type composer -> open /workspace -> Cancel -> draft text still present

# UI: chips / recent do NOT commit
tap Quick Home | Recent row -> browser path changes
  -> GET status workspace UNCHANGED until Use this folder

# UI: commit
browse to dir -> Use this folder -> PUT + navigate / -> home label matches

# UI: cancel
Cancel -> home; no PUT side effect

# UI: browse + files visibility
enter dir updates path; dirs always listed (incl. dot dirs)
  -> files hidden by default; workspace-show-files expands/collapses
  -> files non-selectable when shown; re-hide on path change
  -> omit show-files control when zero files
  -> document order: dirs… → workspace-show-files → files… (when expanded)
```

**Selectors / contracts**

| Selector / field | Role |
|------------------|------|
| `[data-testid="workspace"]` | Home path control; opens `/workspace` |
| `[data-testid="workspace-label"]` | Visible path text on home |
| `[data-testid="workspace-selector"]` | Selector page root |
| `[data-testid="workspace-quick-home"]` | Quick Home chip (browse only) |
| `[data-testid="workspace-quick-cwd"]` | Quick Server cwd (browse only) |
| `[data-testid="workspace-recent-item"]` | Recent row (browse only) |
| `[data-testid="workspace-browser-path"]` | Current browse path text |
| `[data-testid="workspace-browser-parent"]` | Parent directory |
| `[data-testid="workspace-browser-entry"]` | Dir (enter) or file (disabled; hidden until expand) |
| `[data-testid="workspace-show-files"]` | Expand/collapse files (`aria-expanded`); after last dir, before files |
| `[data-testid="workspace-use-folder"]` | **Only** commit CTA |
| `[data-testid="workspace-cancel"]` | Cancel back to home |
| `[data-testid="composer-input"]` | Home draft input (app-level) |

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web-workspace-selector/
├── DOCTEST.md
├── SETUP.md                              # build agent-run, free port, HTTP + playwright helpers
├── api/                                  # Mode=http — multi-step API probes
│   ├── SETUP.md
│   ├── status/                           # GET /status resolve & persist
│   │   ├── SETUP.md
│   │   ├── fresh-defaults-to-cwd/        # A1: no selected → workspace ≈ cwd
│   │   └── persists-after-put/           # A4: second GET keeps selected   [RED]
│   ├── put/                              # PUT /workspace
│   │   ├── SETUP.md
│   │   ├── valid-dir-sets-selected-mru/  # A2: 200 + selected + MRU head  [RED]
│   │   ├── invalid/
│   │   │   ├── SETUP.md
│   │   │   ├── missing-path-400/         # A3: missing path → 400         [RED]
│   │   │   └── file-path-400/            # A3: file path → 400            [RED]
│   │   └── mru-move-to-front-and-cap/    # A7: reorder + cap ≤12          [RED]
│   ├── sessions/
│   │   ├── SETUP.md
│   │   └── new-uses-selected-workspace/  # A5: POST meta.workspace        [RED]
│   └── fs-list/
│       ├── SETUP.md
│       ├── dirs-and-files-typed/              # A6: dir+file typed entries
│       └── dirs-before-files-include-dots/    # sort dirs first + include .dots  [RED]
└── ui/                                   # Mode=ui — Playwright (label ui-automation)
    ├── SETUP.md
    ├── open-selector/                    # home → /workspace
    ├── draft-survives-roundtrip/         # composer survives Cancel
    ├── browse-only/                      # chips/recent never auto-commit
    │   ├── SETUP.md
    │   ├── quick-home-no-commit/
    │   ├── quick-cwd-no-commit/
    │   └── recent-row-no-commit/
    ├── use-this-folder-commits/          # Use → PUT + home
    ├── cancel-no-commit/                 # Cancel no PUT
    ├── browse-enter-dir-files-disabled/  # expand files then assert disabled  [UPDATED]
    ├── files-hidden-by-default/          # dirs only until expand             [RED]
    ├── show-files-toggle/                # expand/collapse + disabled files   [RED]
    ├── show-files-reset-on-path-change/  # re-hide files after enter dir      [RED]
    ├── browse-dot-dir-enterable/         # .hidden-dir listed + enterable     [RED]
    ├── show-files-control-hidden-when-no-files/  # omit toggle if 0 files    [RED]
    └── show-files-toggle-after-last-dir/ # DOM order dirs→toggle→files        [RED]
```

Parameter ranking (most → least significant):

1. **Assertion surface** — HTTP API vs Playwright UI (`Mode=http|ui`).
2. **API resource** — status resolve | put workspace | create session | fs list.
3. **PUT outcome** — valid commit | invalid reject | MRU reorder/cap.
4. **Invalid kind** — missing path vs existing file.
5. **FS list contract** — typed entries | sort + include dots.
6. **UI action class** — open | draft | browse-only | commit Use | cancel | browser nav | files visibility.
7. **Browse-only input** — Quick Home chip vs Recent row (both non-commit).
8. **Files visibility** — default hide | toggle | reset on path | omit when empty | dot dir enter | toggle after last dir.

## Test Index

| # | Leaf | Description | Expect RED today? |
|---|------|-------------|-------------------|
| 1 | `api/status/fresh-defaults-to-cwd` | No `selected_workspace`; `workspace` ≈ process cwd; status exposes `process_cwd` / `home` / `recent_workspaces` | No (baseline) |
| 2 | `api/status/persists-after-put` | PUT valid dir then GET status still selected | No (baseline) |
| 3 | `api/put/valid-dir-sets-selected-mru` | PUT dir → 200; selected set; first in recent; config.json | No (baseline) |
| 4 | `api/put/invalid/missing-path-400` | PUT missing path → 400; no config change | No (baseline) |
| 5 | `api/put/invalid/file-path-400` | PUT existing file → 400; no config change | No (baseline) |
| 6 | `api/put/mru-move-to-front-and-cap` | Re-select moves to front; MRU length ≤ 12 | No (baseline) |
| 7 | `api/sessions/new-uses-selected-workspace` | After PUT, POST sessions → `session.workspace` = selected | No (baseline) |
| 8 | `api/fs-list/dirs-and-files-typed` | List fixture tree: dirs enterable type, files typed as file | No (baseline) |
| 9 | `api/fs-list/dirs-before-files-include-dots` | Dirs before files; include `.git` / `.env`; case-insensitive within group | **Yes** (dots skipped; no dirs-first sort) |
| 10 | `ui/open-selector` | Home workspace control → URL `/workspace`, selector visible | No (baseline) |
| 11 | `ui/draft-survives-roundtrip` | Type draft → open selector → Cancel → draft remains | No (baseline) |
| 12 | `ui/browse-only/quick-home-no-commit` | Quick Home changes browse path; status workspace unchanged | No (baseline) |
| 13 | `ui/browse-only/quick-cwd-no-commit` | Quick Server cwd browse only; status workspace unchanged | No (baseline) |
| 14 | `ui/browse-only/recent-row-no-commit` | Recent row browse only; status workspace unchanged | No (baseline) |
| 15 | `ui/use-this-folder-commits` | Browse + Use → home; workspace label matches selected | No (baseline) |
| 16 | `ui/cancel-no-commit` | Cancel returns home without changing selected workspace | No (baseline) |
| 17 | `ui/browse-enter-dir-files-disabled` | Expand show-files then assert file disabled; enter dir | **Yes** (needs expand + `workspace-show-files`) |
| 18 | `ui/files-hidden-by-default` | Dirs visible; no file rows; show-files collapsed | **Yes** (files currently always shown) |
| 19 | `ui/show-files-toggle` | Expand reveals files (incl. `.env`); collapse hides; files non-selectable | **Yes** |
| 20 | `ui/show-files-reset-on-path-change` | After expand + enter dir, files re-hidden | **Yes** |
| 21 | `ui/browse-dot-dir-enterable` | `.hidden-dir` listed and enterable | **Yes** (API skips dots) |
| 22 | `ui/show-files-control-hidden-when-no-files` | No `workspace-show-files` when listing has only dirs | **Yes** |
| 23 | `ui/show-files-toggle-after-last-dir` | Document order: dirs → `workspace-show-files` → files (collapsed + expanded) | **Yes** (toggle currently above list) |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/web-workspace-selector                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/web-workspace-selector
doctest test --label-all ./cmd/agent-run/tests/web-workspace-selector

doctest vet ./cmd/agent-run/tests/web-workspace-selector
doctest test ./cmd/agent-run/tests/web-workspace-selector
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/api
doctest test --label ui-automation ./cmd/agent-run/tests/web-workspace-selector
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/api/fs-list/dirs-before-files-include-dots
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/ui/files-hidden-by-default --label ui-automation
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/ui/show-files-toggle --label ui-automation
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/ui/show-files-toggle-after-last-dir --label ui-automation
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/ui/browse-enter-dir-files-disabled --label ui-automation
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/ui/use-this-folder-commits --label ui-automation
doctest test -v ./cmd/agent-run/tests/web-workspace-selector/api/put/valid-dir-sets-selected-mru
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

// HTTPStep is one request in a multi-step API scenario.
type HTTPStep struct {
	Name   string // optional label for Assert lookup
	Method string // GET, PUT, POST; default GET
	Path   string // absolute path on server, e.g. /api/agent-run/status
	Body   string // raw JSON body for PUT/POST
	Auth   string // "bearer" | "none" | "" (default: bearer when Token set)
}

type Request struct {
	TempDir  string
	Home     string
	RepoRoot string
	AgentRun string
	Port     int
	Token    string
	// WebTokenMode: "omit" | "explicit" | "auto" (default explicit)
	WebTokenMode string
	BaseURL      string
	Env          []string

	// Mode: "http" | "ui"
	Mode string
	// Scenario slug for assert cross-checks
	Scenario string

	// Optional process cwd for agent-run web (status process_cwd / default workspace)
	WebWorkingDir string

	// Paths under test
	SelectPath  string // absolute directory to select via PUT / Use
	FixtureRoot string // absolute directory tree for fs/list
	OSUserHome  string // expected Quick Home / status.home

	// HTTP multi-step (Mode=http)
	HTTPSteps []HTTPStep

	// Playwright (Mode=ui)
	PlaywrightScript string

	webCmd *exec.Cmd
}

type HTTPResult struct {
	Name        string
	Method      string
	Path        string
	Status      int
	Body        string
	ContentType string
}

type Response struct {
	HTTPResults []HTTPResult
	HTTPStatus  int
	HTTPBody    string
	// ConfigAfter is raw AGENT_RUN_HOME/config.json after HTTP steps (may be empty).
	ConfigAfter string

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

func doHTTPStep(t *testing.T, req *Request, step HTTPStep) HTTPResult {
	t.Helper()
	method := strings.ToUpper(strings.TrimSpace(step.Method))
	if method == "" {
		method = http.MethodGet
	}
	url := strings.TrimRight(req.BaseURL, "/") + step.Path
	var bodyReader io.Reader
	if step.Body != "" {
		bodyReader = strings.NewReader(step.Body)
	}
	httpReq, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, step.Path, err)
	}
	if step.Body != "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	bearer := ""
	switch step.Auth {
	case "none":
		bearer = ""
	case "bearer":
		bearer = req.Token
	default:
		if req.Token != "" && req.WebTokenMode != "omit" {
			bearer = req.Token
		}
	}
	if bearer != "" {
		httpReq.Header.Set("Authorization", "Bearer "+bearer)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return HTTPResult{
		Name:        step.Name,
		Method:      method,
		Path:        step.Path,
		Status:      resp.StatusCode,
		Body:        string(body),
		ContentType: resp.Header.Get("Content-Type"),
	}
}

func runHTTPSteps(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL empty — leaf Setup must start web")
	}
	if len(req.HTTPSteps) == 0 {
		return nil, fmt.Errorf("HTTPSteps empty — leaf Setup must set steps")
	}
	var results []HTTPResult
	for _, step := range req.HTTPSteps {
		results = append(results, doHTTPStep(t, req, step))
	}
	last := results[len(results)-1]
	resp := &Response{
		HTTPResults: results,
		HTTPStatus:  last.Status,
		HTTPBody:    last.Body,
	}
	cfgPath := filepath.Join(req.Home, "config.json")
	if b, err := os.ReadFile(cfgPath); err == nil {
		resp.ConfigAfter = string(b)
	}
	return resp, nil
}

func runUIProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	requirePlaywright(t)
	if strings.TrimSpace(req.PlaywrightScript) == "" {
		return nil, fmt.Errorf("PlaywrightScript is empty — leaf Setup must set it")
	}
	stdout, stderr, exitCode, err := runPlaywrightScript(t, req.PlaywrightScript)
	return &Response{
		PlaywrightStdout: stdout,
		PlaywrightStderr: stderr,
		PlaywrightExit:   exitCode,
		Err:              err,
	}, err
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.webCmd == nil && req.Port > 0 {
		if err := startWebBackground(t, req); err != nil {
			return nil, err
		}
	}
	switch req.Mode {
	case "http", "":
		if req.Mode == "" {
			req.Mode = "http"
		}
		return runHTTPSteps(t, req)
	case "ui":
		return runUIProbe(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q (want http|ui)", req.Mode)
	}
}
```
