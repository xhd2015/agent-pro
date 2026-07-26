# agentrunbridge — `Run` + `RunInteractiveOpen` + `RunInteractiveDetach`

Doc-style tests for `github.com/xhd2015/agent-pro/pkgs/agentrunbridge`, a shared
library that maps structured options to `agent-run` CLI argv, optionally polls
`tty status` until ready, and exposes thin interactive profiles:
`--open` (`RunInteractiveOpen`) and `--detach` (`RunInteractiveDetach`).

# DSN (Domain Specific Notion)

**Participants**

- **`Run`** — full-config entrypoint. Maps `RunOpts` → argv, resolves binary,
  invokes `agent-run` via hooks or real exec, optionally waits until session ready.
- **`RunInteractiveOpen`** — minimal interactive **open** profile. Fills assumed
  `RunOpts` flags and **only** calls `Run` (no second exec path).
- **`RunInteractiveDetach`** — minimal interactive **detach** profile. Fills
  assumed `RunOpts` flags and **only** calls `Run` (same single exec path).
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
- Empty `AgentRunner` in `Run` omits `--agent-runner`; interactive profiles
  (`RunInteractiveOpen` / `RunInteractiveDetach`) default it to `grok-tty`.
- Interactive **open** mapping: `AutoSendOrResume`, `NewTerminal`, `Open`,
  `WaitReady` true; `CaptureStdout` false; `Detach` false.
- Interactive **detach** mapping: `AutoSendOrResume`, `Detach`, `WaitReady`
  true; `NewTerminal`, `Open`, `CaptureStdout` false.
- `Detach` on `RunOpts` → argv token `--detach`. Preferred mutual exclusion:
  `Open` and `Detach` must not both be true (implementer may error in `Run` or
  refuse to emit both flags).
- When `Detach` is true, prompt placement matches open profile: `--` separator
  then prompt (SeaTalk parity for detach local-bot logs).
- `Env []string` on `RunOpts` / `InteractiveOpenOpts`: each `"KEY=VALUE"` emits
  repeatable `-e KEY=VALUE` (two argv tokens: `-e` then `KEY=VALUE`) **after**
  other flags and **before** `--` / prompt.
- Ready when `EqualFold(screen,"banner") && EqualFold(sendable,"yes")`.
- Defaults: `ReadyTimeout` 0 → 60s; `ReadyPollInterval` 0 → 500ms.
- `CaptureStdout` true → `RunResult.Stdout` is trimmed launch stdout.
- No durable storage; no real agent-run / iTerm in these unit leaves.

**Locked argv form (SeaTalk open parity)**

```text
<binary> run --session-id=<id> --agent-runner=<runner> \
  --auto-send-or-resume --new-terminal [--dir=<ws>] [--no-submit] --open -- <prompt>
```

**Locked argv form (detach profile)**

```text
<binary> run --session-id=<id> --agent-runner=<runner> \
  --auto-send-or-resume [--dir=<ws>] [--allow-relocate-resume-session-dir] \
  --detach -- <prompt>
```

Equals-form flags (`--session-id=value`, `--agent-runner=value`, `--dir=value`).
Detach profile: **no** `--open`, **no** `--new-terminal`.

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
│   ├── allow-relocate-resume-session-dir/ # + allow-relocate flag
│   ├── detach-auto-send/               # Detach+AutoSend; no open/new-terminal
│   ├── detach-dir/                     # detach + dir + allow-relocate
│   ├── keep-tty-session/               # --keep-tty, session, no --open
│   ├── auto-send-no-open/              # --auto-send-or-resume, Open false
│   ├── stateless/                      # run + prompt; no session id
│   ├── runner-empty-omitted/           # empty AgentRunner → no --agent-runner
│   └── env-flags/                      # Env → -e KEY=VALUE before --
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
├── interactive-open/                   # RunInteractiveOpen → fills RunOpts → Run
│   ├── defaults-wait-ready/            # minimal opts + wait success
│   ├── dir-nosubmit/                   # WorkspaceDir + NoSubmit in argv
│   ├── runner-default-grok-tty/        # empty AgentRunner injects grok-tty
│   ├── same-argv-as-run-fill/          # spy: same argv as equivalent RunOpts fill
│   └── env-forwarded/                  # InteractiveOpenOpts.Env → -e in launch argv
└── interactive-detach/                 # RunInteractiveDetach → fills RunOpts → Run
    ├── defaults-wait-ready/            # AutoSend+Detach+WaitReady; no open/new-terminal
    ├── dir-relocate/                   # WorkspaceDir + allow-relocate in argv
    ├── runner-default-grok-tty/        # empty AgentRunner injects grok-tty
    └── same-argv-as-run-fill/          # spy: same argv as filled detach RunOpts
```

Parameter ranking (most → least significant):

1. **API / concern** — pure args, pure status, `Run` exec path, interactive open vs detach profile
2. **Validation vs exec vs wait** — empty prompt / binary / capture / poll outcomes
3. **Flag profile** — open/detach defaults, dir/nosubmit/relocate, keep-tty, stateless, runner defaults, Env

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `build-args/interactive-open-defaults` | Open-profile `RunOpts` → argv with session, grok-tty, auto-send, new-terminal, open, `--`, prompt |
| 2 | `build-args/interactive-open-dir-nosubmit` | Also `--dir=` and `--no-submit` |
| 2a | `build-args/allow-relocate-resume-session-dir` | Open profile + `--allow-relocate-resume-session-dir` |
| 2b | `build-args/detach-auto-send` | Detach profile: auto-send + `--detach`; no `--open` / `--new-terminal` |
| 2c | `build-args/detach-dir` | Detach + `--dir=` + `--allow-relocate-resume-session-dir` |
| 3 | `build-args/keep-tty-session` | `--keep-tty` + session; no `--open` |
| 4 | `build-args/auto-send-no-open` | `--auto-send-or-resume` present; no `--open` |
| 5 | `build-args/stateless` | `run` + prompt only; no session id flag |
| 6 | `build-args/runner-empty-omitted` | Empty `AgentRunner` omits `--agent-runner` |
| 6a | `build-args/env-flags` | `Env` → `-e KEY=VALUE` after flags, before `--` prompt |
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
| 19 | `interactive-open/env-forwarded` | `InteractiveOpenOpts.Env` forwarded as `-e` in launch argv |
| 20 | `interactive-detach/defaults-wait-ready` | Minimal InteractiveDetach + wait; detach argv |
| 21 | `interactive-detach/dir-relocate` | Dir + allow-relocate in detach launch argv |
| 22 | `interactive-detach/runner-default-grok-tty` | Empty runner → `--agent-runner=grok-tty` |
| 23 | `interactive-detach/same-argv-as-run-fill` | Launch argv equals `BuildArgs` of filled detach `RunOpts` |

## How to Run

```sh
doctest vet ./tests/agentrunbridge
doctest test ./tests/agentrunbridge

doctest test -v ./tests/agentrunbridge/build-args/interactive-open-defaults
doctest test -v ./tests/agentrunbridge/build-args/detach-auto-send
doctest test -v ./tests/agentrunbridge/build-args/detach-dir
doctest test -v ./tests/agentrunbridge/run/wait-ready-timeout
doctest test -v ./tests/agentrunbridge/interactive-open/same-argv-as-run-fill
doctest test -v ./tests/agentrunbridge/interactive-detach/defaults-wait-ready
doctest test -v ./tests/agentrunbridge/interactive-detach/same-argv-as-run-fill
doctest test -v ./tests/agentrunbridge/build-args/env-flags
doctest test -v ./tests/agentrunbridge/interactive-open/env-forwarded
```

```go
import (

	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunbridge"
	"github.com/xhd2015/doctest/session"
)

// Request drives one leaf via Mode. Leaves set Mode in Setup (root→leaf chain).
type Request struct {
	Mode string // build_args | status | run | interactive_open | interactive_open_vs_run | interactive_detach | interactive_detach_vs_run

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
	// Detach maps to agent-run --detach (RunInteractiveDetach profile).
	Detach   bool
	NoSubmit bool
	// AllowRelocateResumeSessionDir maps to agent-run --allow-relocate-resume-session-dir.
	AllowRelocateResumeSessionDir bool
	Stateless                     bool
	WaitReady                     bool
	ReadyTimeout                  time.Duration
	ReadyPollInterval             time.Duration
	CaptureStdout                 bool
	Env                           []string // each "KEY=VALUE" → agent-run -e KEY=VALUE

	// status mode
	StatusStdout string

	// Scripted hooks for run / interactive_* modes.
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

	// ExpectedArgs is filled in interactive_*_vs_run (BuildArgs of filled RunOpts).
	ExpectedArgs []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
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
		return runPackage(t, req, resp, "run")

	case "interactive_open":
		return runPackage(t, req, resp, "interactive_open")

	case "interactive_open_vs_run":
		// Capture launch argv from RunInteractiveOpen, compare to BuildArgs of filled opts.
		filled := filledInteractiveRunOpts(req)
		resp.ExpectedArgs = agentrunbridge.BuildArgs(filled)
		if _, err := runPackage(t, req, resp, "interactive_open"); err != nil && resp.ErrString == "" {
			// runPackage already stores API errors on resp; only surface harness errors.
			return resp, err
		}
		return resp, nil

	case "interactive_detach":
		return runPackage(t, req, resp, "interactive_detach")

	case "interactive_detach_vs_run":
		filled := filledInteractiveDetachRunOpts(req)
		resp.ExpectedArgs = agentrunbridge.BuildArgs(filled)
		if _, err := runPackage(t, req, resp, "interactive_detach"); err != nil && resp.ErrString == "" {
			return resp, err
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

// hookBag holds recorded exec traffic for one Run / interactive profile call.
type hookBag struct {
	mu sync.Mutex

	lookPathCalls int
	launchCalls   int
	statusCalls   int

	binaryLookedUp string
	launchBinary   string
	launchArgs     []string
}

// profile selects which package entrypoint to call.
// "run" | "interactive_open" | "interactive_detach"
func runPackage(t *testing.T, req *Request, resp *Response, profile string) (*Response, error) {
	t.Helper()
	rec := &hookBag{}
	optsHooks := makeHooks(req, rec)

	var (
		result agentrunbridge.RunResult
		err    error
	)
	switch profile {
	case "interactive_open":
		result, err = agentrunbridge.RunInteractiveOpen(toInteractiveOpts(req, &optsHooks))
	case "interactive_detach":
		// Classic RED until RunInteractiveDetach exists.
		result, err = agentrunbridge.RunInteractiveDetach(toInteractiveOpts(req, &optsHooks))
	default:
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
		Prompt:                        req.Prompt,
		SessionID:                     req.SessionID,
		Binary:                        req.Binary,
		AgentRunner:                   req.AgentRunner,
		RunnerConfigHome:              req.RunnerConfigHome,
		WorkspaceDir:                  req.WorkspaceDir,
		AutoSendOrResume:              req.AutoSendOrResume,
		KeepTTY:                       req.KeepTTY,
		NewTerminal:                   req.NewTerminal,
		Open:                          req.Open,
		Detach:                        req.Detach,
		NoSubmit:                      req.NoSubmit,
		AllowRelocateResumeSessionDir: req.AllowRelocateResumeSessionDir,
		Stateless:                     req.Stateless,
		WaitReady:                     req.WaitReady,
		ReadyTimeout:                  req.ReadyTimeout,
		ReadyPollInterval:             req.ReadyPollInterval,
		CaptureStdout:                 req.CaptureStdout,
		Env:                           append([]string(nil), req.Env...),
	}
	if hooks != nil {
		opts.LookPath = hooks.LookPath
		opts.RunCommand = hooks.RunCommand
		opts.RunOutput = hooks.RunOutput
	}
	return opts
}

func toInteractiveOpts(req *Request, hooks *hookFns) agentrunbridge.InteractiveOpenOpts {
	// InteractiveOpenOpts is shared by RunInteractiveOpen and RunInteractiveDetach
	// (minimal session/prompt/dir hooks). Profile fill differs inside each wrapper.
	opts := agentrunbridge.InteractiveOpenOpts{
		SessionID:                     req.SessionID,
		Prompt:                        req.Prompt,
		WorkspaceDir:                  req.WorkspaceDir,
		NoSubmit:                      req.NoSubmit,
		AllowRelocateResumeSessionDir: req.AllowRelocateResumeSessionDir,
		Binary:                        req.Binary,
		AgentRunner:                   req.AgentRunner,
		Env:                           append([]string(nil), req.Env...),
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
		Prompt:                        req.Prompt,
		SessionID:                     req.SessionID,
		Binary:                        req.Binary,
		AgentRunner:                   runner,
		WorkspaceDir:                  req.WorkspaceDir,
		NoSubmit:                      req.NoSubmit,
		AllowRelocateResumeSessionDir: req.AllowRelocateResumeSessionDir,
		AutoSendOrResume:              true,
		NewTerminal:                   true,
		Open:                          true,
		Detach:                        false,
		WaitReady:                     true,
		CaptureStdout:                 false,
		Env:                           append([]string(nil), req.Env...),
	}
}

// filledInteractiveDetachRunOpts mirrors RunInteractiveDetach → RunOpts mapping.
func filledInteractiveDetachRunOpts(req *Request) agentrunbridge.RunOpts {
	runner := strings.TrimSpace(req.AgentRunner)
	if runner == "" {
		runner = "grok-tty"
	}
	return agentrunbridge.RunOpts{
		Prompt:                        req.Prompt,
		SessionID:                     req.SessionID,
		Binary:                        req.Binary,
		AgentRunner:                   runner,
		WorkspaceDir:                  req.WorkspaceDir,
		NoSubmit:                      req.NoSubmit,
		AllowRelocateResumeSessionDir: req.AllowRelocateResumeSessionDir,
		AutoSendOrResume:              true,
		NewTerminal:                   false,
		Open:                          false,
		Detach:                        true,
		WaitReady:                     true,
		CaptureStdout:                 false,
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
