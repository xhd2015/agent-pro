# Web Sessions List — pagination, search, newest-first

Doctests for the agent-run web **home sessions list** optimization:

1. **Server query API** — `GET /api/agent-run/sessions?limit=&offset=&q=&status=`
2. **Newest-first** order (`updated_at` desc, then `created_at` desc)
3. **UI** — search box, page size 30 + load more, list ordered newest at top

Reuses harness patterns from `web-workspace-selector` / `web-layout`
(`startWebBackground`, free port, multi-step HTTP, `playwright-debug`,
**flat** session layout `sessions/<session_id>/meta.json`).

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web** — binds `127.0.0.1:<port>`; serves SPA + `/api/agent-run/*`.
  Isolated per-leaf via `AGENT_RUN_HOME`.
- **Session store** — flat dirs `AGENT_RUN_HOME/sessions/<session_id>/meta.json`
  with `session_id`, `runner`, `status`, `initial_prompt`, `workspace`,
  `created_at`, `updated_at`.
- **Sessions list API** — `GET /api/agent-run/sessions`:
  - Query: `limit`, `offset`, `q`, `status` (`all` default | `running` | `done`
    where `done` = finished + idle).
  - Omit `limit` → **all** matching sessions (backward compat); `has_more=false`;
    `limit` may be 0 or total.
  - With `limit` → page of sessions; response includes `total`, `limit`,
    `offset`, `has_more`, `counts`.
  - Sort: newest first (`updated_at` desc, then `created_at` desc).
  - `q`: case-insensitive substring on `initial_prompt`, `session_id`,
    `workspace`, `runner`.
  - `total` respects `q` + `status`; **`counts` ignore `q`** (status buckets
    over the full store so chips stay stable while typing).
- **HomePage** — loads first page (`limit=30&offset=0`), search
  (`session-search`), status chips using server `counts`, **explicit** load-more
  only via `[data-testid="session-load-more"]` (button lives **inside** the
  scrollable `[data-testid="session-list"]` at list end; scrolling near bottom
  does **not** auto-fetch). Poll refreshes the loaded window
  (`limit=min(loadedCount,150)`, `offset=0`) on an adaptive sessions-only
  cadence (does not infinite-scroll).
- **Test harness** — build binary, seed metas, free port, HTTP steps and/or
  Playwright at 390×844.

**Behaviors**

```
# Compat: no limit (today's shape + newest-first preferred)
GET /api/agent-run/sessions
  -> sessions: all matching, newest first
  -> has_more=false when field present; limit 0|total

# Paginate
GET ?limit=2&offset=0 -> 2 newest; total=N; has_more=(N>2)
GET ?limit=2&offset=2 -> next page; no session_id overlap with page0

# Search q
GET ?q=TOKEN -> only sessions matching prompt|id|workspace|runner
  -> total = match count

# Status filter
GET ?status=running -> only running; done = finished|idle

# Counts (stable under q)
GET ?q=TOKEN -> counts.all/running/done still full-store buckets
  while total reflects q

# UI newest-first
seed controlled updated_at -> home session-list first row = newest

# UI search
type session-search -> list shows only matches

# UI load more (explicit button only — no infinite scroll)
seed >30 -> initial ≤30 rows; [data-testid=session-load-more] inside session-list
  -> scroll near bottom alone does not append
  -> click session-load-more appends when has_more
```

**Selectors / contracts**

| Selector / field | Role |
|------------------|------|
| `GET /api/agent-run/sessions` | List + optional pagination/filter |
| `sessions[]` | Page (or all if no limit), newest first |
| `total` | Match count after q+status |
| `limit` / `offset` / `has_more` | Pagination |
| `counts.{all,running,done}` | Chip totals; independent of `q` |
| `[data-testid="session-list"]` | Scrollable home list (contains load-more footer) |
| `[data-testid="session-item"]` | One session row |
| `[data-testid="session-preview"]` | Prompt / id label |
| `[data-testid="session-search"]` | Home search input |
| `[data-testid="session-load-more"]` | Explicit load-more **button** at list end (not auto-scroll) |
| `[data-testid="session-filter-all\|running\|done"]` | Status chips |

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web-sessions-list/
├── DOCTEST.md
├── SETUP.md                         # build agent-run, free port, seed helpers, HTTP + playwright
├── api/                             # Mode=http — GET sessions contracts
│   ├── SETUP.md
│   ├── no-limit-compat/             # omit limit → all + newest first
│   ├── limit-offset-page/           # limit=2 pages; no overlap; total/has_more
│   ├── q-filters/                   # q matches prompt / id / workspace / runner
│   ├── status-running/              # status=running only
│   └── counts-present/              # counts ignore q; total applies q
└── ui/                              # Mode=ui — Playwright (label ui-automation)
    ├── SETUP.md
    ├── newest-first-visible/        # first session-item is newest
    ├── search-filters-list/         # session-search narrows list
    ├── load-more-inside-list/       # load-more button inside session-list scroll content
    ├── scroll-does-not-auto-load/   # scroll near bottom does not append
    └── load-more-appends/           # page size 30; button click appends (no scroll fallback)
```

Parameter ranking (most → least significant):

1. **Assertion surface** — HTTP API vs Playwright UI (`Mode=http|ui`).
2. **API contract dimension** — compat (no limit) | pagination | text `q` | status | counts.
3. **UI capability** — sort order | search | load-more placement | no auto-load | button append.

## Test Index

| # | Leaf | Description | Expect RED today? |
|---|------|-------------|-------------------|
| 1 | `api/no-limit-compat` | Omit limit → all N sessions; first is newest by `updated_at` | **Yes** (unsorted / no newest-first on GET) |
| 2 | `api/limit-offset-page` | `limit=2` pages; total; has_more; no id overlap | **Yes** (limit/offset ignored) |
| 3 | `api/q-filters` | `q` filters prompt/id/workspace/runner; total = matches | **Yes** (`q` ignored) |
| 4 | `api/status-running` | `status=running` returns only running | **Yes** (`status` ignored) |
| 5 | `api/counts-present` | `counts` present; stable under `q`; `total` uses `q` | **Yes** (no counts field) |
| 6 | `ui/newest-first-visible` | First `session-item` is newest seed | No (product GREEN) |
| 7 | `ui/search-filters-list` | `session-search` narrows visible items | No (product GREEN) |
| 8 | `ui/load-more-inside-list` | `session-load-more` contained by `session-list` (list-end, not sticky sibling) | No (product GREEN) |
| 9 | `ui/scroll-does-not-auto-load` | Scroll to bottom without click keeps first-page count (~30) | No (product GREEN) |
| 10 | `ui/load-more-appends` | Initial ≤30; **button** click appends when has_more | No (product GREEN) |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/web-sessions-list                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/web-sessions-list
doctest test --label-all ./cmd/agent-run/tests/web-sessions-list

doctest vet ./cmd/agent-run/tests/web-sessions-list
doctest test ./cmd/agent-run/tests/web-sessions-list
doctest test -v ./cmd/agent-run/tests/web-sessions-list/api
doctest test --label ui-automation ./cmd/agent-run/tests/web-sessions-list
doctest test -v ./cmd/agent-run/tests/web-sessions-list/api/limit-offset-page
doctest test -v ./cmd/agent-run/tests/web-sessions-list/ui/newest-first-visible --label ui-automation
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
	Path   string // absolute path on server, e.g. /api/agent-run/sessions?limit=2
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

	// Seeded session specs (filled by leaf Setup via helpers)
	Seeded []SeedSession

	// HTTP multi-step (Mode=http)
	HTTPSteps []HTTPStep

	// Playwright (Mode=ui)
	PlaywrightScript string

	webCmd *exec.Cmd
}

// SeedSession describes one flat meta.json under AGENT_RUN_HOME/sessions/<id>/.
type SeedSession struct {
	SessionID     string
	Runner        string
	Status        string
	InitialPrompt string
	Workspace     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
	return &Response{
		HTTPResults: results,
		HTTPStatus:  last.Status,
		HTTPBody:    last.Body,
	}, nil
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
