# Codex TTY Update Modal Skip Protocol

End-to-end and parse tests for `agent/codex/tty` FetchStatus auto-Skip on the
blocking **Update available** menu, plus `ParseStatusSnapshot` on signed status chrome.

Fixtures + fake TUI scripts: `testdata/update-modal-skip/` (copied from ai-critic
signed captures; see `PROTOCOL.md`).

# DSN (Domain Specific Notion)

**Participants**

- **FetchStatus (`agent/codex/tty`)** — launches ephemeral tty-watch session with
  Codex (or `CODEX_SHOW_STATUS_COMMAND` fake), `waitForPrompt` auto-Skips blocking
  menu (CSI Down `\x1b[B` → verify Skip → Enter `\r` → poll until menu gone), then
  sends `/status` and parses usage fields.
- **ParseStatusSnapshot** — extracts monthly/credits/reset from snapshot text
  (comma-stripped credit amounts).
- **Fake Codex TUI scripts** — `fake-tui-auto-skip.py` / `fake-tui-stuck-update-now.py`
  driven via `CODEX_SHOW_STATUS_COMMAND`.
- **Signed fixtures** — `05-status-fields.snapshot.txt` for offline parse.

**Behaviors**

- Auto-Skip fake: FetchStatus returns usage (monthly 58%, credits 6519/11250, reset).
- Stuck Update now fake: FetchStatus errors; never ready; never Enter while Update now.
- Fixture `05`: ParseStatusSnapshot → 68%, 7698, 11250, `08:00 on 1 Aug`.

## Version

0.0.2

## Decision Tree

```
agent/codex/tty/tests/update-modal-skip/
├── DOCTEST.md
├── SETUP.md
├── testdata/update-modal-skip/       # COPIED fixtures + fake TUI
├── parse/
│    └── status-fields/               # 05 offline ParseStatusSnapshot
└── fetch/
     ├── auto-skip-ready/             # fake menu → Skip → ready usage
     └── stuck-on-update-now/         # refuse Skip → error (negative)
```

Parameter ranking (most → least significant):

1. **Op class** — offline parse vs live FetchStatus
2. **Fake mode / fixture** — auto-skip vs stuck; status screen
3. **Outcome** — ready fields vs error without illegal Enter

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `parse/status-fields` | `05` → ParseStatusSnapshot monthly/credits/reset |
| 2 | `fetch/auto-skip-ready` | fake modal auto-Skip → FetchStatus usage ready |
| 3 | `fetch/stuck-on-update-now` | never leave Update now → error (`slow && negative`) |

## Run profiles (labels)

| Label | Meaning |
|-------|---------|
| `slow` | May wait full timeout until early Skip failure |
| `negative` | Expects error / non-ready |

## How to Run

```sh
# from agent-pro module root
doctest vet ./agent/codex/tty/tests/update-modal-skip
doctest test ./agent/codex/tty/tests/update-modal-skip/...
# skip slow negative by default if using labels:
# doctest test --label '!slow' ./agent/codex/tty/tests/update-modal-skip/...
```

```go
import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	codextty "github.com/xhd2015/agent-pro/agent/codex/tty"
)

// Request drives parse + fetch leaves for update-modal-skip protocol.
type Request struct {
	Op string // parse | fetch

	// parse
	FixtureFile string
	FixturesDir string

	// fetch
	ShowStatusCommand string
	TTYWatchHome      string
	SessionID         string
	FetchTimeoutSecs  int
	StripDaemonPATH   bool
	MarkerDir         string // FAKE_CODEX_MARKER_DIR for negative Enter detection
}

type Response struct {
	// parse / fetch fields
	MonthlyUsage string
	CreditsUsed  string
	CreditsTotal string
	NextReset    string
	ParseErr     string

	// fetch
	FetchOK    bool
	FetchError string
	MarkerFiles []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch strings.TrimSpace(req.Op) {
	case "parse":
		return runParse(t, req, resp)
	case "fetch":
		return runFetch(t, req, resp)
	default:
		return nil, fmt.Errorf("unknown op %q", req.Op)
	}
}

func fixturesDir(req *Request) string {
	if req != nil && strings.TrimSpace(req.FixturesDir) != "" {
		return req.FixturesDir
	}
	candidates := []string{
		filepath.Join(DOCTEST_ROOT, "testdata", "update-modal-skip"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return filepath.Join(DOCTEST_ROOT, "testdata", "update-modal-skip")
}

func runParse(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	if strings.TrimSpace(req.FixtureFile) == "" {
		return nil, fmt.Errorf("FixtureFile required for parse")
	}
	path := filepath.Join(fixturesDir(req), req.FixtureFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	info, parseErr := codextty.ParseStatusSnapshot(string(raw))
	if parseErr != nil {
		resp.ParseErr = parseErr.Error()
		return resp, nil
	}
	if info != nil {
		resp.MonthlyUsage = info.MonthlyUsage
		resp.CreditsUsed = info.CreditsUsed
		resp.CreditsTotal = info.CreditsTotal
		resp.NextReset = info.NextReset
	}
	return resp, nil
}

func runFetch(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	restore := applyFetchEnv(t, req)
	defer restore()

	timeout := req.FetchTimeoutSecs
	if timeout <= 0 {
		timeout = 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	info, err := codextty.FetchStatus(ctx)
	if err != nil {
		resp.FetchOK = false
		resp.FetchError = err.Error()
	} else {
		resp.FetchOK = true
		if info != nil {
			resp.MonthlyUsage = info.MonthlyUsage
			resp.CreditsUsed = info.CreditsUsed
			resp.CreditsTotal = info.CreditsTotal
			resp.NextReset = info.NextReset
		}
	}

	if req.MarkerDir != "" {
		entries, _ := os.ReadDir(req.MarkerDir)
		for _, e := range entries {
			if !e.IsDir() {
				resp.MarkerFiles = append(resp.MarkerFiles, e.Name())
			}
		}
	}
	return resp, nil
}

func applyFetchEnv(t *testing.T, req *Request) func() {
	t.Helper()
	keys := []string{
		"PATH",
		"TTY_WATCH_HOME",
		"CODEX_SHOW_STATUS_COMMAND",
		"CODEX_SHOW_STATUS_SESSION_ID",
		"CODEX_SHOW_STATUS_TIMEOUT",
		"FAKE_CODEX_MARKER_DIR",
	}
	prev := make(map[string]string, len(keys))
	for _, k := range keys {
		prev[k] = os.Getenv(k)
	}

	if req.StripDaemonPATH {
		_ = os.Setenv("PATH", "/usr/bin:/bin:/usr/sbin:/sbin")
	}
	ttyHome := req.TTYWatchHome
	if ttyHome == "" {
		ttyHome = filepath.Join(t.TempDir(), ".tty-watch")
	}
	_ = os.Setenv("TTY_WATCH_HOME", ttyHome)

	showCmd := strings.TrimSpace(req.ShowStatusCommand)
	if showCmd == "" {
		showCmd = autoSkipFakeCommand(req)
	}
	_ = os.Setenv("CODEX_SHOW_STATUS_COMMAND", showCmd)

	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		sid = "codex-update-modal-skip"
	}
	_ = os.Setenv("CODEX_SHOW_STATUS_SESSION_ID", sid)

	timeout := req.FetchTimeoutSecs
	if timeout <= 0 {
		timeout = 30
	}
	_ = os.Setenv("CODEX_SHOW_STATUS_TIMEOUT", strconv.Itoa(timeout))

	if req.MarkerDir != "" {
		_ = os.MkdirAll(req.MarkerDir, 0o755)
		_ = os.Setenv("FAKE_CODEX_MARKER_DIR", req.MarkerDir)
	}

	return func() {
		for _, k := range keys {
			if v, ok := prev[k]; ok {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

func autoSkipFakeCommand(req *Request) string {
	return fakePythonCommand(req, "fake-tui-auto-skip.py")
}

func stuckUpdateNowFakeCommand(req *Request) string {
	return fakePythonCommand(req, "fake-tui-stuck-update-now.py")
}

func fakePythonCommand(req *Request, scriptName string) string {
	script := filepath.Join(fixturesDir(req), scriptName)
	if abs, err := filepath.Abs(script); err == nil {
		script = abs
	}
	python := "python3"
	if p, err := exec.LookPath("python3"); err == nil {
		python = p
	}
	return python + " " + script
}
```
