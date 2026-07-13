# Web SPA + Structured App Tests

Doctests for hardening `agent-run web` as a React Router SPA: server static
fallback (HTML shell + optional session bootstrap), client-side navigation
without full document reloads, NotFound page, and soft auth (no hard reload).

Fine-grained frontend module structure is enforced by implementation review and
frontend vitest (timeline extract) — see **Complementary tests** below. This
tree owns the **new** SPA / nav / NotFound / soft-auth scenarios; existing
`web-layout/` leaves remain regression-only.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web** — binds `127.0.0.1:<port>`; serves embedded SPA from
  `frontend-agent-run/dist` and `/api/agent-run/*` JSON APIs.
- **Static SPA handler (`registerStatic`)** — for non-`/api/` paths: real dist
  assets when present; otherwise fall back to `index.html`. Session page paths
  may inject bootstrap JSON.
- **Session store** — flat layout under `AGENT_RUN_HOME/sessions/<session_id>/`
  (`meta.json`, `events.jsonl`). Bootstrap only when path is exactly
  `/sessions/{session_id}` and the session exists.
- **React SPA (`frontend-agent-run`)** — `BrowserRouter` routes: `/` → HomePage,
  `/sessions/:sessionId` → SessionPage, `*` → NotFoundPage. Auth is a **gate UI**
  (not a dedicated URL).
- **Auth gate** — explicit `--token` mode: empty `localStorage['agent-run-token']`
  shows auth page; submit stores token and **soft-navigates** (no
  `window.location.href` full reload).
- **Test harness** — builds `agent-run`, isolated temp `AGENT_RUN_HOME`, free
  port, `startWebBackground`, HTTP probes and/or `playwright-debug`.

**Behaviors**

```
# SPA shell for app routes
Browser -> GET / | GET /sessions/:id -> agent-run web static -> 200 HTML with #root

# Bootstrap when session exists on valid path shape
Browser -> GET /sessions/<seeded_id> -> static + store.GetSession -> inject
  #agent-run-session-bootstrap JSON containing session_id

# API is not SPA HTML
Browser -> GET /api/agent-run/... -> API mux (not index.html SPA success body)

# Client nav without full document reload
User click session-item | back | not-found-home | auth submit
  -> React Router / soft gate (window marker survives)
```

**Selectors / contracts**

| Selector / signal | Role |
|-------------------|------|
| `#root` | SPA mount node in `index.html` |
| `#agent-run-session-bootstrap` | Optional JSON bootstrap on session HTML |
| `[data-testid="session-item"]` | Home session row link → `/sessions/:id` |
| `[data-testid="chat-active"]` | Session detail main panel |
| `[data-testid="home-active"]` / `[data-testid="empty-state"]` | Home main / empty |
| `[data-testid="auth-page"]` / `[data-testid="auth-token-input"]` | Auth gate |
| `[data-testid="not-found"]` | Unknown client route page |
| `[data-testid="not-found-home"]` | Link/button back to `/` |
| `.back-link` | Session page “← Sessions” control → `/` |
| `window.__SPA_NAV_MARKER` | Test-only marker: survives soft nav; cleared by full reload |

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web-spa-structured-app/
├── DOCTEST.md
├── SETUP.md                              # build agent-run, free port, startWeb, seed helpers
├── spa-static/                           # Mode=http — server SPA fallback
│   ├── SETUP.md
│   ├── root-html-has-root/               # G1: GET / → 200 HTML #root
│   ├── session-path/
│   │   ├── SETUP.md
│   │   ├── unknown-id-spa-shell/         # G2: unknown id → 200 SPA, no bootstrap
│   │   └── seeded-bootstrap/             # G3+G5 accept: bootstrap matches session_id
│   ├── api-not-spa-html/                 # G4: /api/* not SPA HTML shell
│   └── path-parse-rejects-wrong-shapes/  # G5 reject: wrong shapes → no bootstrap
└── client-nav/                           # Mode=ui — Playwright client routing
    ├── SETUP.md
    ├── home-to-session-no-reload/        # P1: home → session, marker survives
    ├── session-back-to-home/             # P2: session → home via back-link
    ├── not-found-page/                   # P3: unknown route UI + home link
    └── soft-auth-no-reload/              # P4: token submit soft auth, marker survives
```

Parameter ranking (most → least significant):

1. **Assertion surface** — HTTP static handler vs Playwright client DOM/nav.
2. **Path class** (static) — `/` vs `/sessions/:id` vs `/api/*` vs wrong-shape paths.
3. **Session existence** (session path) — unknown id vs seeded (bootstrap).
4. **Client nav action** (UI) — home→session vs back vs NotFound vs soft auth.
5. **Token mode** — open API vs explicit Bearer (default explicit for UI; open ok for pure static shell).

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `spa-static/root-html-has-root` | `GET /` → 200, HTML body contains `#root` |
| 2 | `spa-static/session-path/unknown-id-spa-shell` | `GET /sessions/<unknown>` → 200 SPA shell, no bootstrap script |
| 3 | `spa-static/session-path/seeded-bootstrap` | Seeded session → HTML includes bootstrap JSON with matching `session_id` |
| 4 | `spa-static/api-not-spa-html` | `GET /api/agent-run/health` is not successful SPA HTML (`#root` / bootstrap) |
| 5 | `spa-static/path-parse-rejects-wrong-shapes` | Wrong path shapes never inject bootstrap (even with seeded session) |
| 6 | `client-nav/home-to-session-no-reload` | Click `[data-testid="session-item"]` → `/sessions/<id>` + chat-active; no full reload |
| 7 | `client-nav/session-back-to-home` | Session `.back-link` → `/` with home UI |
| 8 | `client-nav/not-found-page` | `/this-route-does-not-exist` → not-found; home control → `/` |
| 9 | `client-nav/soft-auth-no-reload` | Explicit token; submit auth → home; no full document reload |

## Complementary tests (not in this tree)

| Surface | Location | Covers |
|---------|----------|--------|
| Vitest timeline extract | `frontend-agent-run` (`lib/timeline.ts` + `*.test.ts` after split) | V1: merge/dedupe/coalesce/buildTimeline parity with prior App logic |
| Layout / SSE / scroll regression | `cmd/agent-run/tests/web-layout/` | Existing mobile layout leaves stay green |
| Optional Go unit | `cmd/agent-run/web_static*_test.go` | Can unit-test `parseSessionPagePath` directly; G5 is already covered behaviorally via bootstrap leaves above |

```sh
cd frontend-agent-run && bun test   # or npm test — after timeline extract exists
```

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/web-spa-structured-app
doctest test ./cmd/agent-run/tests/web-spa-structured-app
doctest test -v ./cmd/agent-run/tests/web-spa-structured-app/spa-static
doctest test -v ./cmd/agent-run/tests/web-spa-structured-app/client-nav
doctest test --label ui-automation ./cmd/agent-run/tests/web-spa-structured-app
doctest test -v ./cmd/agent-run/tests/web-spa-structured-app/client-nav/home-to-session-no-reload --label ui-automation
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
	"strings"
	"testing"
	"time"
)

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
	// Scenario slug for assert cross-checks (optional)
	Scenario string

	// HTTP probe (Mode=http)
	HTTPMethod string   // default GET
	HTTPPath   string   // single path
	HTTPAuth   string   // "bearer" | "none" | "" (default bearer when Token set for API paths)
	HTTPPaths  []string // multi-path probes; if non-empty, overrides HTTPPath

	// Seed
	SessionID string

	// Playwright (Mode=ui)
	PlaywrightScript string

	webCmd *exec.Cmd
}

type HTTPResult struct {
	Path        string
	Status      int
	Body        string
	ContentType string
}

type Response struct {
	HTTPStatus      int
	HTTPBody        string
	HTTPContentType string
	HTTPResults     []HTTPResult

	PlaywrightStdout string
	PlaywrightStderr string
	PlaywrightExit   int

	Err error
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

func httpGetPath(t *testing.T, baseURL, path, bearer string) HTTPResult {
	t.Helper()
	url := strings.TrimRight(baseURL, "/") + path
	client := &http.Client{Timeout: 10 * time.Second}
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if bearer != "" {
		httpReq.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return HTTPResult{
		Path:        path,
		Status:      resp.StatusCode,
		Body:        string(body),
		ContentType: resp.Header.Get("Content-Type"),
	}
}

func runHTTPProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL empty — leaf Setup must start web")
	}
	paths := req.HTTPPaths
	if len(paths) == 0 {
		if strings.TrimSpace(req.HTTPPath) == "" {
			return nil, fmt.Errorf("HTTPPath/HTTPPaths empty")
		}
		paths = []string{req.HTTPPath}
	}
	var results []HTTPResult
	for _, p := range paths {
		bearer := ""
		switch req.HTTPAuth {
		case "none":
			bearer = ""
		case "bearer":
			bearer = req.Token
		default:
			// SPA paths need no auth; API probes often need bearer when token mode is explicit.
			if strings.HasPrefix(p, "/api/") && req.Token != "" && req.WebTokenMode != "omit" {
				bearer = req.Token
			}
		}
		// Allow explicit none even for API (e.g. expect 401 without SPA HTML).
		if req.HTTPAuth == "none" {
			bearer = ""
		}
		results = append(results, httpGetPath(t, req.BaseURL, p, bearer))
	}
	resp := &Response{HTTPResults: results}
	if len(results) == 1 {
		resp.HTTPStatus = results[0].Status
		resp.HTTPBody = results[0].Body
		resp.HTTPContentType = results[0].ContentType
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

func Run(t *testing.T, req *Request) (*Response, error) {
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
		return runHTTPProbe(t, req)
	case "ui":
		return runUIProbe(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q (want http|ui)", req.Mode)
	}
}
```
