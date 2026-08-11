# agent-run model / reasoning-effort pass-through (P1)

Classic TDD (plan **P1 of 2**): agent-pro / `agent-run run` supports **opt-in
pass-through** of model and reasoning effort to ForceNew follow-up children and
to library `Opts`. **No agent-pro defaults** when empty (do not invent
`gpt-5.6-luna` or `max`).

**RED** until implementer:

1. CLI: `agent-run run --model-reasoning-effort LEVEL` (opt-in; omitted/empty → no effort).
2. `FollowUpOpts` gains `Model` + `ModelReasoningEffort`; `BuildFollowUpCommand`
   emits `--model` / `--model-reasoning-effort` **only when non-empty**.
3. ForceNew path (`openAutoInNewTerminal` / callers of `BuildFollowUpCommand`)
   forwards CLI/opts model + effort into `FollowUpOpts`.
4. Empty stays empty on library `Opts` (regression keep) and on follow-up argv.

**Out of scope (P1):** spl `serve --shared` wire (P2); agent-pro product defaults
to luna/max; live CodeLens / e2e; real codex PTY; publishing agent-pro release.

**Related (not this tree):** spl `tests/agent-pro-codex-cli-flags` covers
`ApplyCodexReasoningEffort` / codex argv `-c model_reasoning_effort=…`. Library
`Opts.ModelReasoningEffort` → agenttty already exists; this tree asserts
empty/no-default + CLI/FollowUp surfaces.

# DSN (Domain Specific Notion)

**Participants**

- **`agent-run run` CLI** (`pkgs/agentruncli`) — accepts optional
  `--model-reasoning-effort LEVEL` alongside existing `--model`. Empty/omitted
  → do not invent effort (no default `max`).
- **`agentrunapi.Opts`** — library path: `Model` / `ModelReasoningEffort` when
  non-empty are already plumbed into agentui/agenttty. Empty must remain empty
  through `AutoSendOrResume` dispatch (no invent).
- **`agentrunapi.FollowUpOpts`** — pure ForceNew child builder inputs. New
  fields: `Model`, `ModelReasoningEffort` (strings; trim; empty = omit).
- **`BuildFollowUpCommand`** — pure: opts → single shell-quoted line; never
  includes `--new-terminal`. When Model non-empty → include model flag; when
  Effort non-empty → include reasoning-effort flag; when empty → neither flag
  and no invented values.
- **ForceNew open path** — `openAutoInNewTerminal` (and any production builder
  of FollowUp for re-exec) must copy model/effort from the parent run into
  `FollowUpOpts` so the child re-exec keeps pass-through.
- **Caller (later P2)** — spl localbot / serve `--shared` supplies values; P1
  only ensures **if set, they pass through**.

**Behaviors**

```
# CLI (agent-run run) — pure L2 (source / flag wire; no process stdio swap)
run_cmd.go runHelp + String registration document/accept --model-reasoning-effort
(no flag / empty)                → effort empty; no default max

# BuildFollowUpCommand / FollowUpOpts
Model="o3", Effort="high"        → follow-up contains --model o3 and --model-reasoning-effort high
Model="o3", Effort=""            → --model o3 only; no reasoning flag
Model="", Effort="max"           → --model-reasoning-effort max only; no --model
Model="", Effort=""              → neither flag; no gpt-5.6-luna / no invented max

# Library Opts (keep / no-default)
Opts.ModelReasoningEffort="max"  → RunSession capture sees "max"
Opts.Model="o3"                  → RunSession capture sees "o3"
Opts.Model="", Effort=""         → capture empty; not invent luna/max

# ForceNew source-wire
openAutoInNewTerminal / BuildFollowUpCommand site
  → assigns FollowUpOpts.Model and FollowUpOpts.ModelReasoningEffort
```

**Flag shape (implementer)**

Prefer argv tokens the existing CLI parser already accepts
(`less-flags` `String`):

- `--model <m>` **or** `--model=<m>`
- `--model-reasoning-effort <level>` **or** `--model-reasoning-effort=<level>`

Harness asserts accept either equals or two-token form. Token checks must **not**
treat `--model-reasoning-effort` as a hit for `--model` (prefix-safe).

## Version

0.0.1

## Decision Tree

```text
tests/agent-run-model-reasoning-effort/
├── DOCTEST.md
├── SETUP.md
├── follow-up/                                   # pure BuildFollowUpCommand (L2)
│   ├── SETUP.md
│   ├── both-set/                                # model=o3 + effort=high
│   ├── model-only/                              # model=o3, effort empty
│   ├── effort-only/                             # model empty, effort=max
│   └── both-empty/                              # neither flag; no defaults
├── library-opts/                                # AutoSendOrResume capture (L2)
│   ├── SETUP.md
│   ├── effort-set/                              # ModelReasoningEffort=max on hook opts
│   ├── model-set/                               # Model=o3 on hook opts
│   └── both-empty-no-default/                   # empty stays empty (no luna/max)
├── cli/                                         # agent-run run surface (L2 help + wire)
│   ├── SETUP.md
│   ├── help-lists-effort/                       # run_cmd.go help surface documents flag
│   └── source-wire-flag/                        # pkgs/agentruncli registers flag
└── source-wire/                                 # ForceNew forwards model/effort
    ├── SETUP.md
    └── follow-up-forwards-model-effort/         # open path + FollowUpOpts fields
```

Parameter ranking (most → least significant):

1. **Surface** — follow-up | library-opts | cli | source-wire
2. **Model × Effort presence** — both / model-only / effort-only / empty
3. **Independence** — pure string/opts capture; no iTerm, no real codex, no
   `t.Setenv` / `t.Chdir` / process-global env mutation

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `follow-up/both-set` | `Model=o3`, `Effort=high` → both flags present with values |
| 2 | `follow-up/model-only` | `Model=o3`, effort empty → `--model` only; no reasoning flag |
| 3 | `follow-up/effort-only` | model empty, `Effort=max` → reasoning flag only; no `--model` |
| 4 | `follow-up/both-empty` | both empty → neither flag; no invented luna/max |
| 5 | `library-opts/effort-set` | `Opts.ModelReasoningEffort=max` → RunSession capture equals `max` |
| 6 | `library-opts/model-set` | `Opts.Model=o3` → RunSession capture equals `o3` |
| 7 | `library-opts/both-empty-no-default` | empty Model+Effort → capture empty; not luna/max |
| 8 | `cli/help-lists-effort` | `run_cmd.go` help surface documents `--model-reasoning-effort` |
| 9 | `cli/source-wire-flag` | `pkgs/agentruncli` sources register `--model-reasoning-effort` |
| 10 | `source-wire/follow-up-forwards-model-effort` | ForceNew path wires Model + ModelReasoningEffort into FollowUpOpts |

## How to Run

```sh
# from agent-pro module root (external/agent-pro-master-2026-08-11-1)
GOWORK=off doctest vet ./tests/agent-run-model-reasoning-effort
GOWORK=off doctest test ./tests/agent-run-model-reasoning-effort

GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/follow-up/both-set
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/follow-up/model-only
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/follow-up/effort-only
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/follow-up/both-empty
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/library-opts/effort-set
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/library-opts/model-set
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/library-opts/both-empty-no-default
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/cli/help-lists-effort
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/cli/source-wire-flag
GOWORK=off doctest test -v ./tests/agent-run-model-reasoning-effort/source-wire/follow-up-forwards-model-effort
```

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

Expect **RED** (compile and/or assertion) until implementer lands FollowUpOpts
fields + emission, CLI flag, and ForceNew forward.

**Parallel-safe harness:** no process stdio reassignment. CLI help is a pure
read of `pkgs/agentruncli/run_cmd.go` (holds `runHelp`); registration and
ForceNew forward are pure source scans.

### Planned production surface (implementer; not written by designer)

```go
// package agentrunapi

type FollowUpOpts struct {
	// ...existing fields...

	// Model is optional; when non-empty, BuildFollowUpCommand emits
	// --model <m> (or --model=<m>) on the child run argv.
	Model string
	// ModelReasoningEffort is optional; when non-empty, emits
	// --model-reasoning-effort <level> (or equals form). Empty → omit (no default).
	ModelReasoningEffort string
}

// BuildFollowUpCommand: with other run flags, before -e / ExtraArgs / -- prompt:
//   if m := strings.TrimSpace(opts.Model); m != "" {
//       remainder = append(remainder, "--model", m) // or "--model="+m
//   }
//   if e := strings.TrimSpace(opts.ModelReasoningEffort); e != "" {
//       remainder = append(remainder, "--model-reasoning-effort", e)
//   }
// Never invent model or effort when empty.
```

```go
// package agentruncli — runHeadless / runAutoSendOrResume / openAutoInNewTerminal

// CLI:
//   String("--model-reasoning-effort", &modelReasoningEffort)
// help documents: --model-reasoning-effort LEVEL

// autoSendOrResumeOpts gains modelReasoningEffort string
// apiOpts.ModelReasoningEffort = opts.modelReasoningEffort
// openAutoInNewTerminal FollowUpOpts:
//   Model: opts.model
//   ModelReasoningEffort: opts.modelReasoningEffort
```

```go
import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// Fixtures (locked)
const (
	fixtureModel          = "o3"
	fixtureEffort         = "high"
	fixtureEffortMax      = "max"
	forbiddenDefaultModel = "gpt-5.6-luna"
)

// Request drives one leaf via Mode.
type Request struct {
	// Mode: follow_up | library_opts | cli_help | source_wire
	Mode string

	// follow_up + library_opts shared
	SessionID            string
	Prompt               string
	AgentRunner          string
	Open                 bool
	Detach               bool
	Model                string
	ModelReasoningEffort string

	// library_opts
	Home string // agentstorage file-store home (temp)

	// source_wire: "cli_flag" | "follow_up_forward"
	SourceWireTarget string
}

// Response is the harness observation.
type Response struct {
	// follow_up
	FollowUp  string
	ErrString string

	// library_opts
	CapturedModel  string
	CapturedEffort string
	RunCalls       int

	// cli_help: pure read of pkgs/agentruncli/run_cmd.go (runHelp + flag registration surface)
	Stdout         string
	HelpSourceRead bool

	// source_wire
	ScannedFiles     int
	FlagFound        bool
	ModelFieldFound  bool
	EffortFieldFound bool
	BuildFollowUpRef bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch strings.TrimSpace(req.Mode) {
	case "follow_up":
		return runFollowUp(t, req, resp)
	case "library_opts":
		return runLibraryOpts(t, req, resp)
	case "cli_help":
		return runCLIHelp(t, d, req, resp)
	case "source_wire":
		return runSourceWire(t, d, req, resp)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func runFollowUp(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	// Classic RED until FollowUpOpts.Model / ModelReasoningEffort exist and emit.
	line, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{
		SessionID:            req.SessionID,
		Prompt:               req.Prompt,
		AgentRunner:          req.AgentRunner,
		Open:                 req.Open,
		Detach:               req.Detach,
		Model:                req.Model,
		ModelReasoningEffort: req.ModelReasoningEffort,
	})
	resp.FollowUp = line
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func runLibraryOpts(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	home := req.Home
	if home == "" {
		home = filepath.Join(t.TempDir(), ".agent-run")
	}
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		return nil, fmt.Errorf("NewFileStore: %w", err)
	}
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = "sess-lib-opts"
	}
	prompt := req.Prompt
	if strings.TrimSpace(prompt) == "" {
		prompt = "library opts prompt"
	}

	opts := agentrunapi.Opts{
		SessionID:            sessionID,
		Prompt:               prompt,
		Model:                req.Model,
		ModelReasoningEffort: req.ModelReasoningEffort,
		Store:                store,
		Stdout:               io.Discard,
		Stderr:               io.Discard,
		NewTerminal:          false,
		RunSession: func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
			_ = ctx
			_ = meta
			_ = found
			resp.RunCalls++
			resp.CapturedModel = o.Model
			resp.CapturedEffort = o.ModelReasoningEffort
			return nil
		},
	}

	if err := agentrunapi.AutoSendOrResume(context.Background(), opts); err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

// runCLIHelp reads the run subcommand source (holds runHelp + flag parse).
// Parallel-safe: pure file read only (no process stdio swap / no Handle).
func runCLIHelp(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	_ = req
	root, err := findAgentProModuleRoot(d)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(root, "pkgs", "agentruncli", "run_cmd.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read run_cmd.go: %w", err)
	}
	resp.Stdout = string(data)
	resp.HelpSourceRead = true
	return resp, nil
}

func runSourceWire(t *testing.T, d *session.Doctest, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	root, err := findAgentProModuleRoot(d)
	if err != nil {
		return nil, err
	}
	pkgDir := filepath.Join(root, "pkgs", "agentruncli")
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, fmt.Errorf("read pkgs/agentruncli: %w", err)
	}

	target := strings.TrimSpace(req.SourceWireTarget)
	if target == "" {
		target = "cli_flag"
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(pkgDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		resp.ScannedFiles++
		src := string(data)

		switch target {
		case "cli_flag":
			if strings.Contains(src, "--model-reasoning-effort") {
				resp.FlagFound = true
			}
		case "follow_up_forward":
			if !strings.Contains(src, "BuildFollowUpCommand") {
				continue
			}
			resp.BuildFollowUpRef = true
			if hasFollowUpModelAssignment(src) {
				resp.ModelFieldFound = true
			}
			if hasFollowUpEffortAssignment(src) {
				resp.EffortFieldFound = true
			}
		}
	}
	return resp, nil
}

// hasFollowUpModelAssignment reports a FollowUpOpts Model: key (not ModelReasoning*).
func hasFollowUpModelAssignment(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		trim := strings.TrimSpace(line)
		if strings.Contains(trim, "ModelReasoningEffort") {
			continue
		}
		// composite literal field
		if strings.HasPrefix(trim, "Model:") || strings.Contains(trim, " Model:") {
			return true
		}
	}
	return false
}

func hasFollowUpEffortAssignment(src string) bool {
	return strings.Contains(src, "ModelReasoningEffort:") ||
		strings.Contains(src, "ModelReasoningEffort =")
}

func findAgentProModuleRoot(d *session.Doctest) (string, error) {
	if d == nil || strings.TrimSpace(d.DOCTEST_ROOT) == "" {
		return "", fmt.Errorf("DOCTEST_ROOT unavailable")
	}
	start := d.DOCTEST_ROOT
	for dir := start; ; dir = filepath.Dir(dir) {
		modPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modPath)
		if err == nil {
			if strings.Contains(string(data), "module github.com/xhd2015/agent-pro") {
				return dir, nil
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("agent-pro module root not found above %s", start)
		}
	}
}
```
