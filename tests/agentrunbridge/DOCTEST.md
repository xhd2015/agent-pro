# agentrunbridge — `Run` + `RunInteractiveOpen`

Doc-style tests for `github.com/xhd2015/agent-pro/pkgs/agentrunbridge`, a shared
library that maps structured options to `agent-run` CLI argv, optionally polls
`tty status` until ready, and exposes a thin interactive `--open` profile.

# DSN (Domain Specific Notion)

**Participants**

- **`Run`** — full-config entrypoint. Maps `RunOpts` → argv, resolves binary,
  invokes `agent-run` via hooks or real exec, optionally waits until session ready.
- **`RunInteractiveOpen`** — minimal interactive profile. Fills assumed
  `RunOpts` flags and **only** calls `Run` (no second exec path).
- **`BuildArgs`** — pure function: `RunOpts` → CLI args after the binary name
  (starts with `run`). Preferred for argv leaves without exec.
- **`ParseTTYStatus` / `IsSessionReady`** — pure parse of human
  `agent-run tty status <id>` stdout (`screen status:` + `sendable:` lines).
- **Binary resolver** — empty `Binary` → `"agent-run"`; `LookPath` must succeed
  before launch.
- **Exec hooks** (test doubles) — `LookPath`, `RunCommand` (launch),
  `RunOutput` (status polls and optional stdout capture). When set, replace
  real `exec`.
- **Wait-ready loop** — after launch when `WaitReady`, poll
  `tty status <sessionID>` until `banner` + `sendable yes` or `ReadyTimeout`.

**Behaviors**

- Prompt required non-empty after trim; empty → error, no exec.
- Session mode: `--session-id=<id>` (equals form) when not `Stateless`.
- `Stateless`: `run` + optional runner flags + prompt; no session id.
- Empty `AgentRunner` in `Run` omits `--agent-runner`; `RunInteractiveOpen`
  defaults it to `grok-tty`.
- Interactive open mapping: `AutoSendOrResume`, `NewTerminal`, `Open`,
  `WaitReady` true; `CaptureStdout` false.
- Ready when `EqualFold(screen,"banner") && EqualFold(sendable,"yes")`.
- Defaults: `ReadyTimeout` 0 → 60s; `ReadyPollInterval` 0 → 500ms.
- `CaptureStdout` true → `RunResult.Stdout` is trimmed launch stdout.
- No durable storage; no real agent-run / iTerm in these unit leaves.

**Locked argv form (SeaTalk parity)**

```text
<binary> run --session-id=<id> --agent-runner=<runner> \
  --auto-send-or-resume --new-terminal [--dir=<ws>] [--no-submit] --open -- <prompt>
```

Equals-form flags (`--session-id=value`, `--agent-runner=value`, `--dir=value`).

## Version

0.0.2

## Decision Tree

```
tests/agentrunbridge/
├── DOCTEST.md
├── SETUP.md
├── build-args/                         # pure BuildArgs(RunOpts)
│   ├── interactive-open-defaults/      # open profile flags + grok-tty + prompt
│   ├── interactive-open-dir-nosubmit/  # + --dir + --no-submit
│   ├── keep-tty-session/               # --keep-tty, session, no --open
│   ├── auto-send-no-open/              # --auto-send-or-resume, Open false
│   ├── stateless/                      # run + prompt; no session id
│   └── runner-empty-omitted/           # empty AgentRunner → no --agent-runner
├── status/                             # pure ParseTTYStatus / IsSessionReady
│   ├── ready/                          # banner + sendable yes → ready
│   └── not-ready/                      # starting + sendable no → not ready
├── run/                                # Run() with fake LookPath/RunCommand/RunOutput
│   ├── empty-prompt/                   # trim-empty Prompt → error; no exec
│   ├── missing-binary/                 # LookPath fails
│   ├── binary-default/                 # empty Binary → LookPath("agent-run")
│   ├── capture-stdout/                 # CaptureStdout → Result.Stdout trimmed
│   ├── wait-ready-success/             # not-ready then ready → ok
│   └── wait-ready-timeout/             # always not-ready + short timeout → error
└── interactive-open/                   # RunInteractiveOpen → fills RunOpts → Run
    ├── defaults-wait-ready/            # minimal opts + wait success
    ├── dir-nosubmit/                   # WorkspaceDir + NoSubmit in argv
    ├── runner-default-grok-tty/        # empty AgentRunner injects grok-tty
    └── same-argv-as-run-fill/          # spy: same argv as equivalent RunOpts fill
```

Parameter ranking (most → least significant):

1. **API / concern** — pure args, pure status, `Run` exec path, interactive open profile
2. **Validation vs exec vs wait** — empty prompt / binary / capture / poll outcomes
3. **Flag profile** — open defaults, dir/nosubmit, keep-tty, stateless, runner defaults

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `build-args/interactive-open-defaults` | Open-profile `RunOpts` → argv with session, grok-tty, auto-send, new-terminal, open, `--`, prompt |
| 2 | `build-args/interactive-open-dir-nosubmit` | Also `--dir=` and `--no-submit` |
| 3 | `build-args/keep-tty-session` | `--keep-tty` + session; no `--open` |
| 4 | `build-args/auto-send-no-open` | `--auto-send-or-resume` present; no `--open` |
| 5 | `build-args/stateless` | `run` + prompt only; no session id flag |
| 6 | `build-args/runner-empty-omitted` | Empty `AgentRunner` omits `--agent-runner` |
| 7 | `status/ready` | Fixture banner+yes → `IsSessionReady` true |
| 8 | `status/not-ready` | Fixture starting+no → false |
| 9 | `run/empty-prompt` | Empty/whitespace prompt → error; zero launch calls |
| 10 | `run/missing-binary` | `LookPath` fails → error about not found |
| 11 | `run/binary-default` | Empty `Binary` → `LookPath("agent-run")` |
| 12 | `run/capture-stdout` | `CaptureStdout` → trimmed `RunResult.Stdout` |
| 13 | `run/wait-ready-success` | Status poll sequence not-ready then ready |
| 14 | `run/wait-ready-timeout` | Always not-ready + short timeout → timeout error |
| 15 | `interactive-open/defaults-wait-ready` | Minimal InteractiveOpen + wait succeeds |
| 16 | `interactive-open/dir-nosubmit` | Dir + NoSubmit appear in launch argv |
| 17 | `interactive-open/runner-default-grok-tty` | Empty runner → `--agent-runner=grok-tty` |
| 18 | `interactive-open/same-argv-as-run-fill` | Launch argv equals `BuildArgs` of filled `RunOpts` |

## How to Run

```sh
doctest vet ./tests/agentrunbridge
doctest test ./tests/agentrunbridge

doctest test -v ./tests/agentrunbridge/build-args/interactive-open-defaults
doctest test -v ./tests/agentrunbridge/run/wait-ready-timeout
doctest test -v ./tests/agentrunbridge/interactive-open/same-argv-as-run-fill
```

```go
import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunbridge"
)

// Request drives one leaf via Mode. Leaves set Mode in Setup (root→leaf chain).
type Request struct {
	Mode string // build_args | status | run | interactive_open | interactive_open_vs_run

	// Shared option fields (mapped into RunOpts or InteractiveOpenOpts).
	Prompt           string
	SessionID        string
	Binary           string
	AgentRunner      string
	RunnerConfigHome string
	WorkspaceDir     string
	AutoSendOrResume bool
	KeepTTY          bool
	NewTerminal      bool
	Open             bool
	NoSubmit         bool
	Stateless        bool
	WaitReady        bool
	ReadyTimeout     time.Duration
	ReadyPollInterval time.Duration
	CaptureStdout    bool

	// status mode
	StatusStdout string

	// Scripted hooks for run / interactive_open modes.
	LookPathFail    bool
	LookPathBin     string // returned path when LookPath succeeds; default "/fake/agent-run"
	LaunchStdout    string // stdout returned when CaptureStdout path uses RunOutput for launch
	LaunchErr       string // non-empty → launch returns error
	StatusPollSeq   []string // successive RunOutput results for tty status polls
	StatusPollHold  string   // if non-empty and seq exhausted, always return this
}

// Response is the harness observation after calling package APIs.
type Response struct {
	Args []string // BuildArgs result, or captured primary launch argv (after binary)

	// Resolved / observed binary name passed to LookPath and launch.
	BinaryLookedUp string
	LaunchBinary   string

	Screen   string
	Sendable string
	Ready    bool

	Stdout string // RunResult.Stdout

	ErrString string

	LookPathCalls   int
	LaunchCalls     int
	StatusPollCalls int

	// ExpectedArgs is filled in interactive_open_vs_run (BuildArgs of filled RunOpts).
	ExpectedArgs []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch req.Mode {
	case "build_args":
		args := agentrunbridge.BuildArgs(toRunOpts(req, nil))
		resp.Args = args
		return resp, nil

	case "status":
		screen, sendable := agentrunbridge.ParseTTYStatus(req.StatusStdout)
		resp.Screen = screen
		resp.Sendable = sendable
		resp.Ready = agentrunbridge.IsSessionReady(req.StatusStdout)
		return resp, nil

	case "run":
		return runPackage(t, req, resp, false)

	case "interactive_open":
		return runPackage(t, req, resp, true)

	case "interactive_open_vs_run":
		// Capture launch argv from RunInteractiveOpen, compare to BuildArgs of filled opts.
		filled := filledInteractiveRunOpts(req)
		resp.ExpectedArgs = agentrunbridge.BuildArgs(filled)
		if _, err := runPackage(t, req, resp, true); err != nil && resp.ErrString == "" {
			// runPackage already stores API errors on resp; only surface harness errors.
			return resp, err
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

// hookBag holds recorded exec traffic for one Run / RunInteractiveOpen call.
type hookBag struct {
	mu sync.Mutex

	lookPathCalls int
	launchCalls   int
	statusCalls   int

	binaryLookedUp string
	launchBinary   string
	launchArgs     []string
}

func runPackage(t *testing.T, req *Request, resp *Response, interactive bool) (*Response, error) {
	t.Helper()
	rec := &hookBag{}
	optsHooks := makeHooks(req, rec)

	var (
		result agentrunbridge.RunResult
		err    error
	)
	if interactive {
		result, err = agentrunbridge.RunInteractiveOpen(toInteractiveOpts(req, &optsHooks))
	} else {
		result, err = agentrunbridge.Run(toRunOpts(req, &optsHooks))
	}

	rec.mu.Lock()
	resp.LookPathCalls = rec.lookPathCalls
	resp.LaunchCalls = rec.launchCalls
	resp.StatusPollCalls = rec.statusCalls
	resp.BinaryLookedUp = rec.binaryLookedUp
	resp.LaunchBinary = rec.launchBinary
	resp.Args = append([]string(nil), rec.launchArgs...)
	rec.mu.Unlock()

	resp.Stdout = result.Stdout
	if err != nil {
		resp.ErrString = err.Error()
	}
	// API errors are observations for Assert; harness returns nil error.
	return resp, nil
}

type hookFns struct {
	LookPath   func(file string) (string, error)
	RunCommand func(name string, args ...string) error
	RunOutput  func(name string, args ...string) (string, error)
}

func makeHooks(req *Request, rec *hookBag) hookFns {
	pollIdx := 0
	lookPathBin := req.LookPathBin
	if lookPathBin == "" {
		lookPathBin = "/fake/agent-run"
	}

	return hookFns{
		LookPath: func(file string) (string, error) {
			rec.mu.Lock()
			rec.lookPathCalls++
			rec.binaryLookedUp = file
			rec.mu.Unlock()
			if req.LookPathFail {
				return "", fmt.Errorf("executable file not found in $PATH")
			}
			return lookPathBin, nil
		},
		RunCommand: func(name string, args ...string) error {
			rec.mu.Lock()
			rec.launchCalls++
			rec.launchBinary = name
			rec.launchArgs = append([]string(nil), args...)
			rec.mu.Unlock()
			if req.LaunchErr != "" {
				return fmt.Errorf("%s", req.LaunchErr)
			}
			return nil
		},
		RunOutput: func(name string, args ...string) (string, error) {
			// Distinguish launch-with-capture vs tty status polls.
			isStatus := len(args) >= 2 && args[0] == "tty" && args[1] == "status"
			if !isStatus {
				// CaptureStdout launch path may use RunOutput for the primary command.
				rec.mu.Lock()
				rec.launchCalls++
				rec.launchBinary = name
				rec.launchArgs = append([]string(nil), args...)
				rec.mu.Unlock()
				if req.LaunchErr != "" {
					return "", fmt.Errorf("%s", req.LaunchErr)
				}
				return req.LaunchStdout, nil
			}
			rec.mu.Lock()
			rec.statusCalls++
			rec.mu.Unlock()
			out := req.StatusPollHold
			if pollIdx < len(req.StatusPollSeq) {
				out = req.StatusPollSeq[pollIdx]
				pollIdx++
			}
			if out == "" {
				out = statusNotReadyFixture()
			}
			return out, nil
		},
	}
}

func toRunOpts(req *Request, hooks *hookFns) agentrunbridge.RunOpts {
	opts := agentrunbridge.RunOpts{
		Prompt:            req.Prompt,
		SessionID:         req.SessionID,
		Binary:            req.Binary,
		AgentRunner:       req.AgentRunner,
		RunnerConfigHome:  req.RunnerConfigHome,
		WorkspaceDir:      req.WorkspaceDir,
		AutoSendOrResume:  req.AutoSendOrResume,
		KeepTTY:           req.KeepTTY,
		NewTerminal:       req.NewTerminal,
		Open:              req.Open,
		NoSubmit:          req.NoSubmit,
		Stateless:         req.Stateless,
		WaitReady:         req.WaitReady,
		ReadyTimeout:      req.ReadyTimeout,
		ReadyPollInterval: req.ReadyPollInterval,
		CaptureStdout:     req.CaptureStdout,
	}
	if hooks != nil {
		opts.LookPath = hooks.LookPath
		opts.RunCommand = hooks.RunCommand
		opts.RunOutput = hooks.RunOutput
	}
	return opts
}

func toInteractiveOpts(req *Request, hooks *hookFns) agentrunbridge.InteractiveOpenOpts {
	// InteractiveOpenOpts fields per requirement: SessionID, Prompt, WorkspaceDir,
	// NoSubmit, Binary, AgentRunner, Logf, plus optional test hooks (LookPath,
	// RunCommand, RunOutput) forwarded into the filled RunOpts.
	// ReadyTimeout / ReadyPollInterval live only on RunOpts; InteractiveOpen uses
	// package defaults after fill (0 → 60s / 500ms). Short-timeout wait leaves use Mode=run.
	opts := agentrunbridge.InteractiveOpenOpts{
		SessionID:    req.SessionID,
		Prompt:       req.Prompt,
		WorkspaceDir: req.WorkspaceDir,
		NoSubmit:     req.NoSubmit,
		Binary:       req.Binary,
		AgentRunner:  req.AgentRunner,
	}
	if hooks != nil {
		opts.LookPath = hooks.LookPath
		opts.RunCommand = hooks.RunCommand
		opts.RunOutput = hooks.RunOutput
	}
	return opts
}

// filledInteractiveRunOpts mirrors the assumed InteractiveOpen → RunOpts mapping
// for composition tests (same-argv-as-run-fill).
func filledInteractiveRunOpts(req *Request) agentrunbridge.RunOpts {
	runner := strings.TrimSpace(req.AgentRunner)
	if runner == "" {
		runner = "grok-tty"
	}
	return agentrunbridge.RunOpts{
		Prompt:           req.Prompt,
		SessionID:        req.SessionID,
		Binary:           req.Binary,
		AgentRunner:      runner,
		WorkspaceDir:     req.WorkspaceDir,
		NoSubmit:         req.NoSubmit,
		AutoSendOrResume: true,
		NewTerminal:      true,
		Open:             true,
		WaitReady:        true,
		CaptureStdout:    false,
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
