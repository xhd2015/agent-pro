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
	"github.com/xhd2015/doctest/session"
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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	switch strings.TrimSpace(req.Op) {
	case "parse":
		return runParse(t, d, req, resp)
	case "fetch":
		return runFetch(t, d, req, resp)
	default:
		return nil, fmt.Errorf("unknown op %q", req.Op)
	}
}

func fixturesDir(d *session.Doctest, req *Request) string {
	if req != nil && strings.TrimSpace(req.FixturesDir) != "" {
		return req.FixturesDir
	}
	candidates := []string{
		filepath.Join(d.DOCTEST_ROOT, "testdata", "update-modal-skip"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return filepath.Join(d.DOCTEST_ROOT, "testdata", "update-modal-skip")
}

func runParse(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	if strings.TrimSpace(req.FixtureFile) == "" {
		return nil, fmt.Errorf("FixtureFile required for parse")
	}
	path := filepath.Join(fixturesDir(d, req), req.FixtureFile)
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

func runFetch(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	// Options carry hooks; MarkerDir residual: fake Python script reads
	// FAKE_CODEX_MARKER_DIR from child env — inject via `env KEY=val` prefix
	// on the command string (no parent Setenv). StripDaemonPATH residual uses
	// a narrowed PATH only when required (documented below).
	timeout := req.FetchTimeoutSecs
	if timeout <= 0 {
		timeout = 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	showCmd := strings.TrimSpace(req.ShowStatusCommand)
	if showCmd == "" {
		showCmd = autoSkipFakeCommand(d, req)
	}
	if req.MarkerDir != "" {
		_ = os.MkdirAll(req.MarkerDir, 0o755)
		// Child-only env for the fake TUI process without parent Setenv.
		showCmd = "env FAKE_CODEX_MARKER_DIR=" + shellSingleQuote(req.MarkerDir) + " " + showCmd
	}
	ttyHome := req.TTYWatchHome
	if ttyHome == "" {
		ttyHome = filepath.Join(t.TempDir(), ".tty-watch")
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		sid = "codex-update-modal-skip"
	}

	// Residual PATH mutation only for StripDaemonPATH leaves (daemon PATH may
	// hide node). Prefer Options for everything else.
	if req.StripDaemonPATH {
		prevPATH := os.Getenv("PATH")
		_ = os.Setenv("PATH", "/usr/bin:/bin:/usr/sbin:/sbin")
		defer func() { _ = os.Setenv("PATH", prevPATH) }()
	}

	info, err := codextty.FetchStatusWithOptions(ctx, codextty.Options{
		Command:        showCmd,
		SessionID:      sid,
		TimeoutSeconds: timeout,
		TTYWatchHome:   ttyHome,
	})
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

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func autoSkipFakeCommand(d *session.Doctest, req *Request) string {
	return fakePythonCommand(d, req, "fake-tui-auto-skip.py")
}

func stuckUpdateNowFakeCommand(d *session.Doctest, req *Request) string {
	return fakePythonCommand(d, req, "fake-tui-stuck-update-now.py")
}

func fakePythonCommand(d *session.Doctest, req *Request, scriptName string) string {
	script := filepath.Join(fixturesDir(d, req), scriptName)
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
