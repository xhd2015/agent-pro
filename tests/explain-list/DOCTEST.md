# explain list — session history Doctests

End-to-end doctests for the `explain list` subcommand (`cmd/explain` →
`agent/explain`). Lists recent explain sessions from on-disk storage (newest
first), pretty-prints Q/A cards, supports `--limit` / `--color`, and never
invokes the LLM.

```text
explain list [--limit N] [--color]
```

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — user or doctest harness invoking the `explain` binary with
  `list` (or `list --help` / `list -h`).
- **explain CLI** — `cmd/explain` → `agent/explain.RunExplain`. When the first
  positional is exactly `list`, dispatches to the list handler (no agent start
  / resume). Otherwise keeps existing ask/resume behavior (out of scope here).
- **Session store** — directories under
  `$AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME/sessions/{YYYY-MM-DD-HH-mm-ss}-{slug}-{hash8}/session.data`
  (JSON `SessionData`: `agent_runner`, `model`, `messages[]` with
  `role`/`message`). Session time = dirname timestamp prefix; list order is
  created-desc.
- **Fake agent binary** — path in `EXPLAIN_AGENT_PATH`. List must never execute
  it; leaves point at a failing stub so accidental LLM paths fail loudly.
- **Test harness** — builds `explain` once per `doctest test` session, seeds
  session dirs under an isolated debug config home, runs `explain list …` with
  controlled env, captures stdout/stderr/exit code.

**Behaviors**

- Empty store → friendly message (e.g. `No explain sessions yet.`), exit 0,
  trailing newline on stdout.
- With sessions: title line includes shown count, total count, and limit; cards newest-first
  with index, formatted time `YYYY-MM-DD HH:MM:SS`, agent_runner / model,
  turn count (user messages), then all Q/A lines in message order.
- `--limit N` default 10; `N <= 0` → default 10; cap at 100.
- Print each message body in full (no soft-truncate, no `…`); preserve stored
  newlines/whitespace. Multi-line bodies: first line after `Q  `/`A  `; each
  subsequent non-empty line indented with exactly **6 spaces**; empty segments
  emit pure `\n` (blank vertical spacing, no spaces-only lines).
- Skip corrupt dirs (missing/bad `session.data` or unparseable timestamp)
  silently.
- Color policy: `--color` forces ANSI on (wins over `NO_COLOR` and non-TTY);
  else if `NO_COLOR` non-empty → off; else auto (TTY). Harness uses pipes so
  auto is off. When on: `Q` bold cyan (`\x1b[1;36m`), `A` bold green
  (`\x1b[1;32m`), headers/meta/separators dim (`\x1b[2m`); bodies plain.
- `explain list --help` / `-h` documents `--limit` and `--color`.
- List never starts/resumes an agent (no `EXPLAIN_AGENT_PATH` invocation).
- User-facing stdout always ends with a trailing `\n` after the last content
  line.

## Version

0.0.2

## Decision Tree

```
[explain list]
 |
 +-- empty/
 |    |
 |    +-- no-sessions/                 (LEAF) empty config home → friendly msg, exit 0
 |
 +-- order-and-limit/
 |    |
 |    +-- time-desc/                   (LEAF) older + newer → newer first
 |    +-- default-limit-10/            (LEAF) 12 sessions, no --limit → 10 newest
 |    +-- limit-flag/                  (LEAF) --limit 3 with 5 sessions → 3 newest
 |    +-- limit-zero-defaults/         (LEAF) --limit 0 → default 10 behavior
 |    +-- limit-capped-at-100/         (LEAF) --limit 200 → effective limit 100
 |
 +-- content/
 |    |
 |    +-- qa-pairs/                    (LEAF) multi-turn Q/A + time + runner/model/turns
 |    +-- full-body/                   (LEAF) long body printed in full; no …
 |    +-- multiline-indent/            (LEAF) newlines preserved; 6-space continuations
 |    +-- skip-corrupt/                (LEAF) bad dir ignored; good session listed
 |
 +-- color/
 |    |
 |    +-- force-color/                 (LEAF) --color → ANSI on Q/A labels + dim meta
 |    +-- no-color-env/                (LEAF) NO_COLOR set, no --color → no ANSI
 |    +-- color-overrides-no-color/    (LEAF) --color + NO_COLOR → ANSI still on
 |    +-- plain-no-flag/               (LEAF) no --color, non-TTY → no ANSI
 |
 +-- help/
 |    |
 |    +-- list-help/                   (LEAF) list --help mentions --limit and --color
 |
 +-- dispatch/
      |
      +-- no-llm-on-list/              (LEAF) list does not invoke EXPLAIN_AGENT_PATH
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `empty/no-sessions` | Empty store → friendly empty message, exit 0, trailing `\n` |
| 2 | `order-and-limit/time-desc` | Two sessions → newest first by dirname timestamp |
| 3 | `order-and-limit/default-limit-10` | 12 sessions, default limit → 10 newest shown |
| 4 | `order-and-limit/limit-flag` | `--limit 3` with 5 sessions → 3 newest |
| 5 | `order-and-limit/limit-zero-defaults` | `--limit 0` treated as default 10 |
| 6 | `order-and-limit/limit-capped-at-100` | `--limit 200` capped to effective limit 100 |
| 7 | `content/qa-pairs` | Multi-turn card: time, runner/model, turns, Q/A lines |
| 8 | `content/full-body` | Long message (200 `x`) printed in full; no `…` |
| 9 | `content/multiline-indent` | Newlines preserved; non-empty continuations indent 6 spaces |
| 10 | `content/skip-corrupt` | Dir without valid `session.data` skipped silently |
| 11 | `color/force-color` | `--color` emits bold-cyan Q, bold-green A, dim meta |
| 12 | `color/no-color-env` | `NO_COLOR` set without `--color` → no `\x1b` |
| 13 | `color/color-overrides-no-color` | `--color` + `NO_COLOR` → ANSI still present |
| 14 | `color/plain-no-flag` | No flag, non-TTY harness → no ANSI |
| 15 | `help/list-help` | `explain list --help` mentions `--limit` and `--color` |
| 16 | `dispatch/no-llm-on-list` | List never runs fake agent at `EXPLAIN_AGENT_PATH` |

## How to Run

```sh
doctest vet ./tests/explain-list
doctest test ./tests/explain-list/...
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Msg is one role/message pair written into session.data.
type Msg struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

// SessionSeed describes one on-disk session fixture under
// $ConfigHome/sessions/{DirName}/session.data.
type SessionSeed struct {
	DirName     string // full dirname; must start with YYYY-MM-DD-HH-mm-ss
	AgentRunner string
	Model       string
	Messages    []Msg
}

// Request describes one `explain list` (or list --help) invocation.
// Root Setup fills Bin, ConfigHome, FakeAgentPath. Leaves set Args, Sessions, EnvExtra.
type Request struct {
	Bin           string        // path to built explain binary
	Args          []string      // args after binary, default ["list"]
	ConfigHome    string        // AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME
	FakeAgentPath string        // EXPLAIN_AGENT_PATH (failing stub)
	Sessions      []SessionSeed // written under ConfigHome/sessions before Run
	EnvExtra      []string      // extra KEY=VAL entries (e.g. NO_COLOR=1)
	RepoRoot      string        // module root (informational)
}

// Response is the observable outcome of one explain invocation.
type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Bin == "" {
		t.Fatal("req.Bin not set; root Setup must build explain")
	}
	if req.ConfigHome == "" {
		t.Fatal("req.ConfigHome not set; root Setup must isolate debug config home")
	}
	if len(req.Args) == 0 {
		req.Args = []string{"list"}
	}

	if err := seedSessions(req.ConfigHome, req.Sessions); err != nil {
		return nil, fmt.Errorf("seed sessions: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	cmd.Dir = req.ConfigHome
	cmd.Env = buildExplainEnv(req)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = ee.ExitCode()
			// Non-zero exit is a normal product outcome for some leaves; return resp.
			return resp, nil
		}
		return resp, err
	}
	resp.ExitCode = 0
	return resp, nil
}

// buildExplainEnv builds a controlled environment:
// - strips NO_COLOR / debug-home / EXPLAIN_AGENT_PATH from parent so leaves own them
// - injects isolated ConfigHome + FakeAgentPath
// - appends req.EnvExtra (e.g. NO_COLOR=1)
func buildExplainEnv(req *Request) []string {
	strip := map[string]bool{
		"NO_COLOR": true,
		"AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME": true,
		"EXPLAIN_AGENT_PATH": true,
		// Avoid accidental color forcing from developer shells.
		"FORCE_COLOR":   true,
		"CLICOLOR_FORCE": true,
	}
	var env []string
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		if strip[key] {
			continue
		}
		env = append(env, e)
	}
	env = append(env,
		"AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME="+req.ConfigHome,
		"EXPLAIN_AGENT_PATH="+req.FakeAgentPath,
	)
	env = append(env, req.EnvExtra...)
	return env
}

func seedSessions(configHome string, seeds []SessionSeed) error {
	base := filepath.Join(configHome, "sessions")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}
	for _, s := range seeds {
		if s.DirName == "" {
			return fmt.Errorf("SessionSeed missing DirName")
		}
		dir := filepath.Join(base, s.DirName)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		// Empty Messages + empty runner still writes valid JSON unless DirName-only
		// corrupt fixtures intentionally leave the dir without session.data.
		// Convention: AgentRunner == "" && Model == "" && Messages == nil means
		// "create dir only" (corrupt / skip case). Leaves that need empty-but-valid
		// JSON should set AgentRunner non-empty or Messages non-nil empty slice.
		if s.AgentRunner == "" && s.Model == "" && s.Messages == nil {
			continue
		}
		payload := map[string]interface{}{
			"agent_runner":       s.AgentRunner,
			"model":              s.Model,
			"agent_runners_meta": map[string]interface{}{},
			"messages":           s.Messages,
		}
		if s.Messages == nil {
			payload["messages"] = []Msg{}
		}
		raw, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "session.data"), raw, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// --- session-scoped binary + fake agent cache ---

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "explain-list-doctest-"+DOCTEST_SESSION_ID)
}

func withFileLock(lockPath string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func findModuleRoot() (string, error) {
	start := DOCTEST_ROOT
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "explain")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find module root (go.mod + cmd/explain) above %s", start)
		}
	}
}

func buildExplainOnce(t *testing.T) (string, error) {
	t.Helper()
	cache := sessionCacheDir()
	bin := filepath.Join(cache, "explain")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")

	err := withFileLock(lock, func() error {
		if fileExists(ready) && fileExists(bin) {
			return nil
		}
		repoRoot, err := findModuleRoot()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(cache, 0o755); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-o", bin, "./cmd/explain")
		cmd.Dir = repoRoot
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go build ./cmd/explain: %w\n%s", err, be.String())
		}
		return os.WriteFile(ready, []byte("ok\n"), 0o644)
	})
	if err != nil {
		return "", err
	}
	return bin, nil
}

func ensureFakeAgent(t *testing.T) (string, error) {
	t.Helper()
	cache := sessionCacheDir()
	path := filepath.Join(cache, "fake-opencode")
	lock := filepath.Join(cache, "fake-agent.lock")
	ready := filepath.Join(cache, "fake-agent.ready")

	err := withFileLock(lock, func() error {
		if fileExists(ready) && fileExists(path) {
			return nil
		}
		if err := os.MkdirAll(cache, 0o755); err != nil {
			return err
		}
		// Failing stub: any accidental LLM path surfaces clearly.
		script := "#!/bin/sh\necho FAKE_AGENT_INVOKED >&2\nexit 99\n"
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			return err
		}
		return os.WriteFile(ready, []byte("ok\n"), 0o644)
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// simpleSession is a helper for leaves that need one short Q/A session.
func simpleSession(dirName, runner, model, q, a string) SessionSeed {
	return SessionSeed{
		DirName:     dirName,
		AgentRunner: runner,
		Model:       model,
		Messages: []Msg{
			{Role: "user", Message: q},
			{Role: "assistant", Message: a},
		},
	}
}
```
