# agentrunapi — Classify + AutoSendOrResume library

Classic TDD doctests for plan phase **P1**: extract auto-send-or-resume
**classify + send + run + resume** into importable package
`github.com/xhd2015/agent-pro/pkgs/agentrunapi`.

**P2** (WaitReady + FollowUp Driver) lives in a **nested** root so P1 leaves stay
compile-isolated and GREEN while P2 APIs are RED:

```
tests/agentrunapi/wait-driver/   # own DOCTEST.md — see that tree
```

**Out of scope for this (P1) root:** WaitReady / BuildFollowUpCommand /
OpenInNewTerminal (see nested P2), local-bot, slack-msg. These P1 leaves never
require a real `agent-run` binary, iTerm, or grok.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** (later: `cmd/agent-run run --auto-send-or-resume`, and any in-process
  client) needs to decide MODE and dispatch without shelling out to `agent-run`
  for the in-process path (`NewTerminal=false`).
- **`Classify`** — pure session lifecycle classifier. Inputs: `agentstorage.Store`
  + stable session id + optional **Probe** hook. Outputs: `Mode` (`run`|`send`|
  `resume`), resolved `SessionMeta`, `found`, `error`. Same decision rules as
  current `runAutoSendOrResume` classify in `cmd/agent-run/run_cmd.go`.
- **Probe** — supplies lifecycle facts without real TTY/registry when injected.
  Report fields used for classification: `ResumeReady` (bound+exited) and
  `RunnerExited` (`*bool`: nil unknown, false live, true exited).
- **`LifecycleProbe`** — production TTY/registry/process probe. When
  `Classify` / `Opts.Probe` is **nil**, the package uses **`LifecycleProbe`**
  (not `EmptyProbe`). Without a real TTY registry under a temp store, lifecycle
  is unknown → classify **ModeRun** for a found session (unit-testable contract).
- **`EmptyProbe`** — always unknown lifecycle (`ResumeReady=false`,
  `RunnerExited=nil`) for unit tests that need ModeRun without TTY I/O.
- **`AutoSendOrResume`** — validates opts, calls `Classify`, then dispatches:
  - `send` → live follow-up path (agentsend parity / `SendLive` hook)
  - `resume` → resumeExistingSession semantics (`ResumeSession` hook)
  - `run` → create/run path via agentui (`RunSession` hook)
  When `NewTerminal=true`, P1 may leave ForceNew to CLI; unit leaves set
  `NewTerminal=false` and use hooks.
- **Dispatch hooks** (`RunSession` / `SendLive` / `ResumeSession` on `Opts`) —
  test doubles that record which path ran. Nil → production implementation.
  With hooks set, no `agent-run` binary LookPath is required.
- **Store** — `agentstorage.Store` (file store under temp home in tests). Missing
  session is valid (`found=false` → mode `run`).
- **`cmd/agent-run` (source-wire)** — thin main imports `agentrunapi` (blank
  import OK); production auto path lives in **`pkgs/agentruncli`**.
- **`pkgs/agentruncli` (source-wire)** — run/status paths reference
  `agentrunapi.LifecycleProbe` as the shared production probe.

**Behaviors**

```
# Classify decision (after resolve session by id; missing OK)
found=false                              -> ModeRun
found && ResumeReady                     -> ModeResume
found && RunnerExited!=nil && !*Exited   -> ModeSend
else (unbound / exited unknown / …)      -> ModeRun

# Probe selection
probe == nil                             -> LifecycleProbe (production default)
probe == EmptyProbe                      -> unknown lifecycle → else ModeRun when found
probe == LifecycleProbe (empty store)    -> unknown/not live → ModeRun when found
injected script probe                    -> report drives Mode

# AutoSendOrResume validation (before store/probe side effects)
SessionID empty/whitespace -> error (requires session id); no dispatch hooks
Open && Detach             -> mutual exclusive error; no dispatch hooks

# AutoSendOrResume dispatch (NewTerminal=false)
classify -> switch Mode:
  send   -> SendLive hook or agentsend path
  resume -> ResumeSession hook or resumeExistingSession
  run    -> RunSession hook or agentui.Run
# No agent-run binary required when hooks cover the path.

# CLI wire
agentruncli run_cmd / status -> agentrunapi.LifecycleProbe
```

## Version

0.0.2

## Decision Tree

```
tests/agentrunapi/
├── DOCTEST.md
├── SETUP.md
├── api-surface/                         # package + public symbols
│   ├── SETUP.md
│   └── exports-exist/                   # Mode + Classify + probes + AutoSendOrResume
├── classify/                            # Classify(store, id, probe)
│   ├── SETUP.md
│   ├── missing-session/                 # no meta → run, found=false
│   ├── live/                            # found + exited=false → send
│   ├── resume-ready/                    # found + ResumeReady → resume
│   ├── found-else-run/                  # found + script probe unknown → run
│   ├── empty-probe-found-run/           # found + EmptyProbe → run
│   └── nil-probe-found-run/             # found + nil → LifecycleProbe default → run
├── auto-send-or-resume/                 # AutoSendOrResume(ctx, Opts)
│   ├── SETUP.md
│   ├── validation/                      # gates before dispatch
│   │   ├── SETUP.md
│   │   ├── empty-session-id/            # blank id → error; zero hook calls
│   │   └── open-and-detach-mutex/       # Open+Detach → error; zero hook calls
│   └── dispatch/                        # classify then hook; no binary
│       ├── SETUP.md
│       ├── mode-run-hook/               # missing → RunSession; no LookPath
│       ├── mode-send-hook/              # live → SendLive
│       ├── mode-resume-hook/            # resume-ready → ResumeSession
│       ├── nil-probe-mode-run-hook/     # found + Probe nil → RunSession once
│       └── empty-probe-mode-run-hook/   # found + EmptyProbe → RunSession once
├── source-wire/                         # CLI cutover exit criterion
│   ├── SETUP.md
│   ├── cli-imports-agentrunapi/         # cmd/agent-run sources import package
│   └── cli-uses-lifecycle-probe/        # agentruncli references LifecycleProbe
└── wait-driver/                         # P2 NESTED DOCTEST root (separate Run)
    └── …                                # status | wait-ready | follow-up | open-new-terminal
```

Parameter ranking (most → least significant):

1. **API concern** — surface | classify | auto-send-or-resume | source-wire
2. **Session lifecycle state** (classify / dispatch) — missing | live | resume-ready | else-run
3. **Probe selection** — script | EmptyProbe | nil/LifecycleProbe default
4. **Validation vs dispatch** (auto) — empty session / open+detach vs mode hooks
5. **Binary independence** — NewTerminal=false + hooks never need agent-run LookPath

P2 nested ranking: see `wait-driver/DOCTEST.md`.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `api-surface/exports-exist` | Package exports `Mode`, `Classify`, `AutoSendOrResume`, `Opts`, `ProbeReport`, `LifecycleProbe`, `EmptyProbe` |
| 2 | `classify/missing-session` | Empty store / unknown id → `ModeRun`, `found=false`, no error |
| 3 | `classify/live` | Seeded meta + probe `RunnerExited=false` → `ModeSend`, `found=true` |
| 4 | `classify/resume-ready` | Seeded meta + probe `ResumeReady=true` → `ModeResume`, `found=true` |
| 5 | `classify/found-else-run` | Seeded meta + script probe not ready / exited unknown → `ModeRun`, `found=true` |
| 6 | `classify/empty-probe-found-run` | Seeded meta + `EmptyProbe` → `ModeRun`, `found=true` |
| 7 | `classify/nil-probe-found-run` | Seeded meta + nil probe (→ `LifecycleProbe`) → `ModeRun`, `found=true`; no crash |
| 8 | `auto-send-or-resume/validation/empty-session-id` | Empty SessionID → API error; zero dispatch hooks |
| 9 | `auto-send-or-resume/validation/open-and-detach-mutex` | Open+Detach → mutual exclusive error; zero hooks |
| 10 | `auto-send-or-resume/dispatch/mode-run-hook` | Missing session → `RunSession` once; Mode=run; no binary |
| 11 | `auto-send-or-resume/dispatch/mode-send-hook` | Live probe → `SendLive` once; Mode=send |
| 12 | `auto-send-or-resume/dispatch/mode-resume-hook` | ResumeReady → `ResumeSession` once; Mode=resume |
| 13 | `auto-send-or-resume/dispatch/nil-probe-mode-run-hook` | Seeded + Probe nil → `RunSession` once (not SendLive); no binary |
| 14 | `auto-send-or-resume/dispatch/empty-probe-mode-run-hook` | Seeded + EmptyProbe → `RunSession` once |
| 15 | `source-wire/cli-imports-agentrunapi` | `cmd/agent-run` `.go` sources import `pkgs/agentrunapi` |
| 16 | `source-wire/cli-uses-lifecycle-probe` | `pkgs/agentruncli` sources reference `LifecycleProbe` |

## How to Run

```sh
# from agent-pro module root (or path relative to workspace)
doctest vet ./tests/agentrunapi
doctest test ./tests/agentrunapi

doctest test -v ./tests/agentrunapi/api-surface/exports-exist
doctest test -v ./tests/agentrunapi/classify/missing-session
doctest test -v ./tests/agentrunapi/classify/live
doctest test -v ./tests/agentrunapi/classify/resume-ready
doctest test -v ./tests/agentrunapi/classify/found-else-run
doctest test -v ./tests/agentrunapi/classify/empty-probe-found-run
doctest test -v ./tests/agentrunapi/classify/nil-probe-found-run
doctest test -v ./tests/agentrunapi/auto-send-or-resume/validation/empty-session-id
doctest test -v ./tests/agentrunapi/auto-send-or-resume/validation/open-and-detach-mutex
doctest test -v ./tests/agentrunapi/auto-send-or-resume/dispatch/mode-run-hook
doctest test -v ./tests/agentrunapi/auto-send-or-resume/dispatch/mode-send-hook
doctest test -v ./tests/agentrunapi/auto-send-or-resume/dispatch/mode-resume-hook
doctest test -v ./tests/agentrunapi/auto-send-or-resume/dispatch/nil-probe-mode-run-hook
doctest test -v ./tests/agentrunapi/auto-send-or-resume/dispatch/empty-probe-mode-run-hook
doctest test -v ./tests/agentrunapi/source-wire/cli-imports-agentrunapi
doctest test -v ./tests/agentrunapi/source-wire/cli-uses-lifecycle-probe

# P2 nested tree (independent; WaitReady/FollowUp may be GREEN when implemented)
doctest vet ./tests/agentrunapi/wait-driver
doctest test ./tests/agentrunapi/wait-driver
```

P1 leaves are **GREEN** for Classify/AutoSendOrResume + LifecycleProbe/EmptyProbe
coverage. Existing `cmd/agent-run/tests/auto-send-or-resume` remains the CLI
integration regression suite (not part of this tree).

### Planned public API (RED until implementer)

```go
package agentrunapi

import (
	"context"
	"io"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// Mode is the auto-send-or-resume branch: run | send | resume.
type Mode string

const (
	ModeRun    Mode = "run"
	ModeSend   Mode = "send"
	ModeResume Mode = "resume"
)

// ProbeReport is the lifecycle subset Classify needs (from probeSessionStatus parity).
type ProbeReport struct {
	// ResumeReady is true when runner is bound and exited (resume path).
	ResumeReady bool
	// RunnerExited is nil when unknown; false when live; true when exited.
	RunnerExited *bool
}

// ProbeFunc injects lifecycle probing. nil → LifecycleProbe (not EmptyProbe).
type ProbeFunc func(store agentstorage.Store, meta agentstorage.SessionMeta) (ProbeReport, error)

// LifecycleProbe is the production TTY/registry/process probe (Classify default).
func LifecycleProbe(store agentstorage.Store, meta agentstorage.SessionMeta) (ProbeReport, error)

// EmptyProbe always returns unknown lifecycle for unit ModeRun without TTY I/O.
func EmptyProbe(store agentstorage.Store, meta agentstorage.SessionMeta) (ProbeReport, error)

// Classify resolves session id and returns Mode using the same rules as
// cmd/agent-run runAutoSendOrResume. Missing session → ModeRun, found=false.
func Classify(store agentstorage.Store, sessionID string, probe ProbeFunc) (mode Mode, meta agentstorage.SessionMeta, found bool, err error)

// Opts drives AutoSendOrResume. NewTerminal=false is the in-process P1 path.
// Dispatch hooks, when set, replace production send/run/resume for unit tests
// and prove no agent-run binary LookPath is required.
type Opts struct {
	SessionID                     string
	Prompt                        string
	WorkspaceDir                  string
	AgentRunner                   string
	AgentRunnerBinary             string
	RunnerConfigHome              string
	Model                         string
	Open                          bool
	Detach                        bool
	NoSubmit                      bool
	KeepTTY                       bool
	JSON                          bool
	AllowRelocateResumeSessionDir bool
	// NewTerminal: P1 unit leaves keep false. When true, ForceNew may remain CLI-owned.
	NewTerminal bool
	Env         []string
	PrependPaths []string
	Store       agentstorage.Store
	Stdout      io.Writer
	Stderr      io.Writer
	Probe       ProbeFunc

	// Optional dispatch overrides (nil = production path).
	RunSession    func(ctx context.Context, opts Opts, meta agentstorage.SessionMeta, found bool) error
	SendLive      func(ctx context.Context, opts Opts, meta agentstorage.SessionMeta) error
	ResumeSession func(ctx context.Context, opts Opts, meta agentstorage.SessionMeta) error
}

// AutoSendOrResume validates, classifies, and dispatches run|send|resume.
func AutoSendOrResume(ctx context.Context, opts Opts) error
```

```go
import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// Request drives one leaf via Mode. Leaves set Mode in Setup (root→leaf chain).
type Request struct {
	// Mode: api_surface | classify | auto | source_wire
	Mode string

	SessionID string
	Prompt    string
	Home      string // agentstorage file-store home (temp)
	// SeedMeta writes a minimal session under Home before classify/auto.
	SeedMeta          bool
	Runner            string
	RunnerSessionID   string
	TerminalSessionID string
	Workspace         string
	MetaStatus        string

	// Probe script for classify / auto (nil fields = defaults per leaf Setup).
	UseProbe     bool
	ResumeReady  bool
	// RunnerExited: "live" | "exited" | "unknown" (empty = unknown)
	RunnerExited string
	// ProbeName selects probe for classify/auto:
	//   "" | "nil" → nil (production LifecycleProbe inside Classify), unless UseProbe
	//   "empty" → EmptyProbe
	//   "lifecycle" → LifecycleProbe (explicit)
	//   "script" → makeProbe (same as UseProbe)
	ProbeName string

	// AutoSendOrResume option fields
	Open   bool
	Detach bool
	// InstallHooks records dispatch without production send/run/resume.
	InstallHooks bool
	// Fail if any hook is invoked (validation leaves).
	ExpectNoHooks bool

	// SourceWireTarget: "cmd" (default) scans cmd/agent-run for import;
	// "agentruncli" scans pkgs/agentruncli for LifecycleProbe symbol.
	SourceWireTarget string
}

// Response is the harness observation after calling package APIs.
type Response struct {
	Mode      agentrunapi.Mode
	Found     bool
	MetaID    string
	ErrString string

	// Dispatch hook call counts (auto mode with InstallHooks).
	RunCalls    int
	SendCalls   int
	ResumeCalls int
	HookMode    agentrunapi.Mode // mode observed when first hook fired

	// source_wire
	ImportFound          bool
	LifecycleProbeFound  bool
	ScannedFiles         int
	// api_surface: EmptyProbe/LifecycleProbe callable
	EmptyProbeOK      bool
	LifecycleProbeOK  bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch req.Mode {
	case "api_surface":
		return runAPISurface(t, d, req, resp)
	case "classify":
		return runClassify(t, d, req, resp)
	case "auto":
		return runAuto(t, d, req, resp)
	case "source_wire":
		return runSourceWire(t, d, req, resp)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func runAPISurface(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = d
	// Touch public symbols so the leaf fails to compile until exports exist.
	_ = agentrunapi.ModeRun
	_ = agentrunapi.ModeSend
	_ = agentrunapi.ModeResume
	_ = agentrunapi.ProbeReport{}
	_ = agentrunapi.Opts{}
	// LifecycleProbe / EmptyProbe must be package-level funcs (callable).
	var _ agentrunapi.ProbeFunc = agentrunapi.LifecycleProbe
	var _ agentrunapi.ProbeFunc = agentrunapi.EmptyProbe
	// Classify / AutoSendOrResume are invoked with empty fixtures; errors OK.
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		return nil, err
	}
	mode, _, found, err := agentrunapi.Classify(store, "api-surface-missing", nil)
	resp.Mode = mode
	resp.Found = found
	if err != nil {
		resp.ErrString = err.Error()
	}
	// EmptyProbe always succeeds with zero report.
	if rep, err := agentrunapi.EmptyProbe(store, agentstorage.SessionMeta{}); err == nil && !rep.ResumeReady && rep.RunnerExited == nil {
		resp.EmptyProbeOK = true
	}
	// LifecycleProbe on empty store/meta must not panic; empty report is OK.
	if _, err := agentrunapi.LifecycleProbe(store, agentstorage.SessionMeta{}); err == nil {
		resp.LifecycleProbeOK = true
	}
	// AutoSendOrResume with empty session must be callable (validation error expected).
	if err := agentrunapi.AutoSendOrResume(context.Background(), agentrunapi.Opts{
		Store: store,
	}); err != nil {
		// prefer keeping first classify err empty; store auto err if classify was fine
		if resp.ErrString == "" {
			resp.ErrString = err.Error()
		}
	}
	return resp, nil
}

func runClassify(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = d
	store, err := openAndMaybeSeed(t, req)
	if err != nil {
		return nil, err
	}
	probe := selectProbe(req)
	mode, meta, found, err := agentrunapi.Classify(store, req.SessionID, probe)
	resp.Mode = mode
	resp.Found = found
	resp.MetaID = meta.SessionID
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func runAuto(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = d
	store, err := openAndMaybeSeed(t, req)
	if err != nil {
		return nil, err
	}

	opts := agentrunapi.Opts{
		SessionID:    req.SessionID,
		Prompt:       req.Prompt,
		Open:         req.Open,
		Detach:       req.Detach,
		Store:        store,
		AgentRunner:  req.Runner,
		WorkspaceDir: req.Workspace,
		// P1 unit path: never new-terminal / never agent-run binary.
		NewTerminal: false,
		Probe:       selectProbe(req),
	}
	// Resume precheck needs on-disk grok updates.jsonl when runner is grok-tty;
	// otherwise AutoSendOrResume clears the bind and ModeRuns (post resume_recover).
	// Use RunnerConfigHome (not GROK_HOME env) so parallel doctest leaves stay safe.
	if home, err := ensureLocalGrokUpdatesForResume(t, req); err != nil {
		return nil, err
	} else if home != "" {
		opts.RunnerConfigHome = home
	}
	if req.InstallHooks {
		opts.RunSession = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
			resp.RunCalls++
			if resp.HookMode == "" {
				resp.HookMode = agentrunapi.ModeRun
			}
			return nil
		}
		opts.SendLive = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
			resp.SendCalls++
			if resp.HookMode == "" {
				resp.HookMode = agentrunapi.ModeSend
			}
			return nil
		}
		opts.ResumeSession = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
			resp.ResumeCalls++
			if resp.HookMode == "" {
				resp.HookMode = agentrunapi.ModeResume
			}
			return nil
		}
	}

	err = agentrunapi.AutoSendOrResume(context.Background(), opts)
	if err != nil {
		resp.ErrString = err.Error()
	}
	// Surface classified mode for dispatch leaves via hook observation.
	if resp.HookMode != "" {
		resp.Mode = resp.HookMode
	}
	return resp, nil
}

func runSourceWire(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	// d.DOCTEST_ROOT is tests/agentrunapi; module root is ../..
	moduleRoot := filepath.Join(d.DOCTEST_ROOT, "..", "..")
	target := strings.TrimSpace(req.SourceWireTarget)
	if target == "" {
		target = "cmd"
	}
	switch target {
	case "agentruncli":
		return scanAgentruncliLifecycleProbe(t, moduleRoot, resp)
	default:
		return scanCmdAgentRunImport(t, moduleRoot, resp)
	}
}

func scanCmdAgentRunImport(t *testing.T, moduleRoot string, resp *Response) (*Response, error) {
	t.Helper()
	cmdDir := filepath.Join(moduleRoot, "cmd", "agent-run")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil, fmt.Errorf("read cmd/agent-run: %w", err)
	}
	fset := token.NewFileSet()
	importPath := "github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		// Skip tests; wire is production main package sources.
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(cmdDir, e.Name())
		resp.ScannedFiles++
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if p == importPath {
				resp.ImportFound = true
				return resp, nil
			}
		}
		// Fallback token search (in case of build tags / partial parse).
		data, _ := os.ReadFile(path)
		if strings.Contains(string(data), importPath) || strings.Contains(string(data), "pkgs/agentrunapi") {
			resp.ImportFound = true
			return resp, nil
		}
	}
	return resp, nil
}

func scanAgentruncliLifecycleProbe(t *testing.T, moduleRoot string, resp *Response) (*Response, error) {
	t.Helper()
	cliDir := filepath.Join(moduleRoot, "pkgs", "agentruncli")
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		return nil, fmt.Errorf("read pkgs/agentruncli: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(cliDir, e.Name())
		resp.ScannedFiles++
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		src := string(data)
		if strings.Contains(src, "LifecycleProbe") {
			resp.LifecycleProbeFound = true
			// Prefer files that also import agentrunapi for confidence.
			if strings.Contains(src, "pkgs/agentrunapi") || strings.Contains(src, "agentrunapi.LifecycleProbe") {
				resp.ImportFound = true
			}
			return resp, nil
		}
	}
	return resp, nil
}

// selectProbe maps Request probe fields to ProbeFunc.
// nil → Classify uses LifecycleProbe (production default, not EmptyProbe).
func selectProbe(req *Request) agentrunapi.ProbeFunc {
	name := strings.TrimSpace(strings.ToLower(req.ProbeName))
	switch name {
	case "empty":
		return agentrunapi.EmptyProbe
	case "lifecycle":
		return agentrunapi.LifecycleProbe
	case "script":
		return makeProbe(req)
	case "nil":
		return nil
	case "":
		if req.UseProbe {
			return makeProbe(req)
		}
		return nil
	default:
		if req.UseProbe {
			return makeProbe(req)
		}
		return nil
	}
}

func openAndMaybeSeed(t *testing.T, req *Request) (agentstorage.Store, error) {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		return nil, err
	}
	if !req.SeedMeta {
		return store, nil
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		return nil, fmt.Errorf("SeedMeta requires SessionID")
	}
	runner := req.Runner
	if runner == "" {
		runner = "grok-tty"
	}
	status := req.MetaStatus
	if status == "" {
		status = "running"
	}
	meta := agentstorage.SessionMeta{
		Runner:            runner,
		SessionID:         sid,
		RunnerSessionID:   req.RunnerSessionID,
		TerminalSessionID: req.TerminalSessionID,
		Status:            status,
		Workspace:         req.Workspace,
	}
	if err := store.CreateSession(sid, meta); err != nil {
		return nil, err
	}
	return store, nil
}

func makeProbe(req *Request) agentrunapi.ProbeFunc {
	return func(store agentstorage.Store, meta agentstorage.SessionMeta) (agentrunapi.ProbeReport, error) {
		rep := agentrunapi.ProbeReport{ResumeReady: req.ResumeReady}
		switch req.RunnerExited {
		case "live":
			f := false
			rep.RunnerExited = &f
		case "exited":
			tr := true
			rep.RunnerExited = &tr
		default:
			// unknown / empty → nil
		}
		return rep, nil
	}
}

// ensureLocalGrokUpdatesForResume writes a minimal updates.jsonl so
// localGrokRunnerSessionMissing is false for grok resume dispatch leaves.
// Returns the grok home to pass as Opts.RunnerConfigHome (parallel-safe; no env).
func ensureLocalGrokUpdatesForResume(t *testing.T, req *Request) (grokHome string, err error) {
	t.Helper()
	runner := strings.ToLower(strings.TrimSpace(req.Runner))
	if runner == "" {
		runner = "grok-tty"
	}
	if !(runner == "grok" || runner == "grok-tty" || strings.HasPrefix(runner, "grok")) {
		return "", nil
	}
	id := strings.TrimSpace(req.RunnerSessionID)
	if id == "" {
		return "", nil
	}
	grokHome = t.TempDir()
	dir := filepath.Join(grokHome, "sessions", "fixture-ws", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte("{}\n"), 0o644); err != nil {
		return "", err
	}
	return grokHome, nil
}
```
