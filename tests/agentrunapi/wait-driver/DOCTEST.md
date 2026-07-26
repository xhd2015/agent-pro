# agentrunapi P2 — WaitReady + FollowUp Driver

Classic TDD doctests for plan phase **P2**: library **WaitReady** (no
`agent-run tty status` binary) and **NewTerminal FollowUp** builder
(`DriverBinary` / `DriverArgsPrefix`, empty driver → `"agent-run"`).

Nested under `tests/agentrunapi/` as its own `DOCTEST.md` root so **P1 leaves
remain GREEN** while these APIs are missing (compile isolation).

**Out of scope:** local-bot, slack-msg, removing agentrunbridge exec path.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** needs session readiness without shelling out, and a shell-quoted
  ForceNew follow-up command that re-invokes auto-send-or-resume **without**
  `--new-terminal` in the child.
- **`IsSessionReadyFromStatus` / `ParseTTYStatus`** — pure parse of human
  `tty status` stdout; ready = EqualFold(screen, `"banner"`) &&
  EqualFold(sendable, `"yes"`) (parity with `agentrunbridge.IsSessionReady`).
  May wrap or re-export bridge helpers.
- **`WaitReady`** — polls injectable `StatusFn` until ready or timeout.
  Defaults: Timeout 0→60s, PollInterval 0→500ms. No binary LookPath.
- **`BuildFollowUpCommand`** — pure: `FollowUpOpts` → single shell-quoted line
  for iTerm `FollowUpCommands`.
  - Empty `DriverBinary` → default `"agent-run"`.
  - Optional `DriverArgsPrefix` tokens after binary, before `run` (spl helper shape).
  - Child argv includes `run` + auto-send flags; **never** emits `--new-terminal`.
  - Open profile: `--auto-send-or-resume`, `--open`, session/runner/dir/nosubmit
    as set; prompt after `--`.
  - Detach profile: `--detach` instead of `--open`; still no `--new-terminal`.
- **`OpenInNewTerminal`** — resolves workspace dir + follow-up string, then
  calls injectable `OpenTerminal(dir, followUp)` (tests) or production iTerm
  ForceNew (implementer). Unit leaves never open real iTerm.
- **`pkgs/agentruncli` (source-wire)** — new-terminal path should call
  `BuildFollowUpCommand` (Driver = process executable when self re-exec).
  Thin `cmd/agent-run` only blank-imports the package.

**Behaviors**

```
# status
tty status stdout -> ParseTTYStatus -> screen, sendable
                 -> IsSessionReadyFromStatus -> ready?

# wait
WaitReady(StatusFn, Timeout, PollInterval)
  poll until ready | timeout error (mentions ready/timeout)

# follow-up (pure)
DriverBinary empty -> "agent-run"
tokens = [driver] + DriverArgsPrefix + [run, flags..., --?, prompt?]
  -> shell-quote each token -> single line
flags: --session-id= --agent-runner= --auto-send-or-resume
       [--dir=] [--no-submit] [--allow-relocate-resume-session-dir]
       --open | --detach
  NEVER --new-terminal

# open new terminal (hook)
OpenInNewTerminal(dir, followUp, OpenTerminal?) -> OpenTerminal(dir, followUp)
```

## Version

0.0.2

## Decision Tree

```
tests/agentrunapi/wait-driver/
├── DOCTEST.md
├── SETUP.md
├── status/                              # pure ParseTTYStatus / IsSessionReadyFromStatus
│   ├── SETUP.md
│   ├── ready/                           # banner + sendable yes → true
│   └── not-ready/                       # starting + no → false
├── wait-ready/                          # WaitReady with StatusFn
│   ├── SETUP.md
│   ├── success-second-poll/             # not-ready then ready
│   ├── timeout/                         # always not-ready + short timeout
│   └── missing-session-id/              # empty SessionID → error; zero polls
├── follow-up/                           # pure BuildFollowUpCommand
│   ├── SETUP.md
│   ├── driver-default-agent-run/        # empty DriverBinary → agent-run
│   ├── custom-driver-prefix/            # binary + ArgsPrefix; no --new-terminal
│   ├── open-child-shape/                # auto-send + open + session; no new-terminal
│   └── detach-child-shape/              # auto-send + detach; no open/new-terminal
├── open-new-terminal/                   # OpenInNewTerminal + injectable open
│   ├── SETUP.md
│   └── injectable-open/                 # records dir + followUp; no real iTerm
└── source-wire/                         # CLI cutover for FollowUp
    ├── SETUP.md
    └── uses-build-follow-up/            # pkgs/agentruncli references BuildFollowUpCommand
```

Parameter ranking (most → least significant):

1. **API concern** — status | wait-ready | follow-up | open-new-terminal | source-wire
2. **Outcome** — ready vs not-ready; success vs timeout vs validation
3. **Driver identity** — default agent-run vs custom binary+prefix
4. **Child flag profile** — open vs detach; never new-terminal

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `status/ready` | Fixture banner+yes → `IsSessionReadyFromStatus` true; screen/sendable parsed |
| 2 | `status/not-ready` | Fixture starting+no → false |
| 3 | `wait-ready/success-second-poll` | StatusFn not-ready then ready → ok; ≥2 polls |
| 4 | `wait-ready/timeout` | Always not-ready + short timeout → error (timeout/ready) |
| 5 | `wait-ready/missing-session-id` | Empty SessionID → error; zero StatusFn calls |
| 6 | `follow-up/driver-default-agent-run` | Empty DriverBinary → line mentions `agent-run` |
| 7 | `follow-up/custom-driver-prefix` | Custom binary + prefix tokens; no `--new-terminal` |
| 8 | `follow-up/open-child-shape` | Session, auto-send, open, `--`, prompt; no `--new-terminal` |
| 9 | `follow-up/detach-child-shape` | Detach profile; no open/new-terminal |
| 10 | `open-new-terminal/injectable-open` | OpenTerminal hook called with dir + follow-up |
| 11 | `source-wire/uses-build-follow-up` | `pkgs/agentruncli` sources reference `BuildFollowUpCommand` |

## How to Run

```sh
# from agent-pro module root
doctest vet ./tests/agentrunapi/wait-driver
doctest test ./tests/agentrunapi/wait-driver

# P1 tree must stay green independently:
doctest test ./tests/agentrunapi/classify
doctest test ./tests/agentrunapi/auto-send-or-resume

doctest test -v ./tests/agentrunapi/wait-driver/status/ready
doctest test -v ./tests/agentrunapi/wait-driver/wait-ready/success-second-poll
doctest test -v ./tests/agentrunapi/wait-driver/wait-ready/timeout
doctest test -v ./tests/agentrunapi/wait-driver/follow-up/driver-default-agent-run
doctest test -v ./tests/agentrunapi/wait-driver/follow-up/custom-driver-prefix
doctest test -v ./tests/agentrunapi/wait-driver/follow-up/open-child-shape
doctest test -v ./tests/agentrunapi/wait-driver/follow-up/detach-child-shape
doctest test -v ./tests/agentrunapi/wait-driver/open-new-terminal/injectable-open
doctest test -v ./tests/agentrunapi/wait-driver/source-wire/uses-build-follow-up
```

Expect **RED** until P2 exports land. P1 leaves under parent `tests/agentrunapi/`
(non-nested) remain GREEN.

### Planned public API (P2 — RED until implementer)

```go
package agentrunapi

import "time"

// ParseTTYStatus extracts screen status and sendable token from human
// `agent-run tty status` stdout (parity with agentrunbridge.ParseTTYStatus).
func ParseTTYStatus(stdout string) (screen, sendable string)

// IsSessionReadyFromStatus reports banner + sendable yes
// (parity with agentrunbridge.IsSessionReady).
func IsSessionReadyFromStatus(stdout string) bool

// WaitReadyOpts polls an injectable status source until the session is ready.
type WaitReadyOpts struct {
	SessionID    string // required (non-empty after trim)
	StatusFn     func() (stdout string, err error) // required
	Timeout      time.Duration // 0 → 60s
	PollInterval time.Duration // 0 → 500ms
}

// WaitReady polls StatusFn until IsSessionReadyFromStatus or timeout.
// Timeout error should mention ready and/or timeout. No agent-run binary.
func WaitReady(opts WaitReadyOpts) error

// FollowUpOpts builds a shell-quoted ForceNew child command (no --new-terminal).
type FollowUpOpts struct {
	// DriverBinary empty → "agent-run". Use os.Executable() for self re-exec.
	DriverBinary string
	// DriverArgsPrefix optional tokens after binary, before "run" (spl helper).
	DriverArgsPrefix []string

	SessionID                     string
	Prompt                        string
	AgentRunner                   string
	WorkspaceDir                  string
	NoSubmit                      bool
	AllowRelocateResumeSessionDir bool
	Open                          bool
	Detach                        bool
	// Env each "KEY=VALUE" → -e KEY=VALUE before -- / prompt (optional).
	Env []string
}

// BuildFollowUpCommand returns a single shell-quoted command line suitable for
// iterm2 FollowUpCommands. Never includes --new-terminal.
// Open/Detach mutual exclusion: if both true, return error.
// Empty SessionID → error.
func BuildFollowUpCommand(opts FollowUpOpts) (string, error)

// OpenInNewTerminalOpts opens a new terminal with a follow-up command.
type OpenInNewTerminalOpts struct {
	WorkspaceDir string
	// FollowUp if non-empty is used as-is; else built from FollowUpOpts.
	FollowUp     string
	FollowUpOpts FollowUpOpts
	// OpenTerminal replaces production iTerm ForceNew when set (unit tests).
	OpenTerminal func(dir string, followUp string) error
}

// OpenInNewTerminal invokes OpenTerminal(dir, followUp). When OpenTerminal is
// nil, implementer may call iterm2 ModeForceNew (not exercised in these units).
func OpenInNewTerminal(opts OpenInNewTerminalOpts) error

// Optional Opts extensions (document for later AutoSendOrResume integration;
// WaitReady may remain a separate call — sealed by wait-ready leaves above):
//   DriverBinary string
//   DriverArgsPrefix []string
```

```go
import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/doctest/session"
)

// Request drives one leaf via Mode.
type Request struct {
	// Mode: status | wait_ready | follow_up | open_new_terminal | source_wire
	Mode string

	// status
	StatusStdout string

	// wait_ready
	SessionID         string
	StatusPollSeq     []string
	StatusPollHold    string // if seq exhausted, keep returning this
	ReadyTimeout      time.Duration
	ReadyPollInterval time.Duration
	// SkipStatusFn when testing missing StatusFn / session validation
	OmitStatusFn bool

	// follow_up / open
	DriverBinary                  string
	DriverArgsPrefix              []string
	Prompt                        string
	AgentRunner                   string
	WorkspaceDir                  string
	NoSubmit                      bool
	AllowRelocateResumeSessionDir bool
	Open                          bool
	Detach                        bool
	Env                           []string

	// open_new_terminal: if FollowUpLine set, use as-is; else build from fields
	FollowUpLine string
}

// Response is the harness observation.
type Response struct {
	Screen   string
	Sendable string
	Ready    bool

	FollowUp  string
	ErrString string

	StatusPollCalls int

	// open_new_terminal hook
	OpenDir      string
	OpenFollowUp string
	OpenCalls    int

	// source_wire
	ImportFound  bool
	SymbolFound  bool
	ScannedFiles int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch req.Mode {
	case "status":
		return runStatus(t, d, req, resp)
	case "wait_ready":
		return runWaitReady(t, d, req, resp)
	case "follow_up":
		return runFollowUp(t, d, req, resp)
	case "open_new_terminal":
		return runOpenNewTerminal(t, d, req, resp)
	case "source_wire":
		return runSourceWire(t, d, req, resp)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func runStatus(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = d
	screen, sendable := agentrunapi.ParseTTYStatus(req.StatusStdout)
	resp.Screen = screen
	resp.Sendable = sendable
	resp.Ready = agentrunapi.IsSessionReadyFromStatus(req.StatusStdout)
	return resp, nil
}

func runWaitReady(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = d
	opts := agentrunapi.WaitReadyOpts{
		SessionID:    req.SessionID,
		Timeout:      req.ReadyTimeout,
		PollInterval: req.ReadyPollInterval,
	}
	if !req.OmitStatusFn {
		pollIdx := 0
		opts.StatusFn = func() (string, error) {
			resp.StatusPollCalls++
			if pollIdx < len(req.StatusPollSeq) {
				out := req.StatusPollSeq[pollIdx]
				pollIdx++
				return out, nil
			}
			if req.StatusPollHold != "" {
				return req.StatusPollHold, nil
			}
			return statusNotReadyFixture(), nil
		}
	}
	err := agentrunapi.WaitReady(opts)
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func runFollowUp(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = d
	line, err := agentrunapi.BuildFollowUpCommand(toFollowUpOpts(req))
	resp.FollowUp = line
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func runOpenNewTerminal(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = d
	opts := agentrunapi.OpenInNewTerminalOpts{
		WorkspaceDir: req.WorkspaceDir,
		FollowUp:     req.FollowUpLine,
		FollowUpOpts: toFollowUpOpts(req),
		OpenTerminal: func(dir, followUp string) error {
			resp.OpenCalls++
			resp.OpenDir = dir
			resp.OpenFollowUp = followUp
			return nil
		},
	}
	err := agentrunapi.OpenInNewTerminal(opts)
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func runSourceWire(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	// Nested root: d.DOCTEST_ROOT = tests/agentrunapi/wait-driver.
	// Production new-terminal FollowUp lives in pkgs/agentruncli (thin cmd/agent-run).
	cliDir := filepath.Join(d.DOCTEST_ROOT, "..", "..", "..", "pkgs", "agentruncli")
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		return nil, fmt.Errorf("read pkgs/agentruncli: %w", err)
	}
	fset := token.NewFileSet()
	importPath := "github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(cliDir, e.Name())
		resp.ScannedFiles++
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		src := string(data)
		if strings.Contains(src, importPath) || strings.Contains(src, "pkgs/agentrunapi") {
			resp.ImportFound = true
		}
		if strings.Contains(src, "BuildFollowUpCommand") {
			resp.SymbolFound = true
		}
		// Also try import parse (best effort)
		if f, err := parser.ParseFile(fset, path, data, parser.ImportsOnly); err == nil {
			for _, imp := range f.Imports {
				if strings.Trim(imp.Path.Value, `"`) == importPath {
					resp.ImportFound = true
				}
			}
		}
	}
	return resp, nil
}

func toFollowUpOpts(req *Request) agentrunapi.FollowUpOpts {
	return agentrunapi.FollowUpOpts{
		DriverBinary:                  req.DriverBinary,
		DriverArgsPrefix:              append([]string(nil), req.DriverArgsPrefix...),
		SessionID:                     req.SessionID,
		Prompt:                        req.Prompt,
		AgentRunner:                   req.AgentRunner,
		WorkspaceDir:                  req.WorkspaceDir,
		NoSubmit:                      req.NoSubmit,
		AllowRelocateResumeSessionDir: req.AllowRelocateResumeSessionDir,
		Open:                          req.Open,
		Detach:                        req.Detach,
		Env:                           append([]string(nil), req.Env...),
	}
}

func statusReadyFixture() string {
	return strings.Join([]string{
		"session: test-sess",
		"screen status: banner",
		"sendable: yes",
		"state: idle",
	}, "\n")
}

func statusNotReadyFixture() string {
	return strings.Join([]string{
		"session: test-sess",
		"screen status: starting",
		"sendable: no",
		"state: booting",
	}, "\n")
}
```
