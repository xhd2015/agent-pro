# agenttty child process env (`BuildChildProcessEnv`)

Plan phase **P3** contract tests: agent-run builds **pure agent argv** and a
child env policy as `Set` / `Unset` slices for tty-watch HeadlessRun
(`CommandEnv` / `CommandUnset`). Never prefix argv with `env(1)`.

Classic TDD: `BuildChildProcessEnv` and `ChildEnvSpec` are **not** implemented
yet. This tree is **RED** (compile-RED for build leaves; assertion-RED for
pure-argv) until the implementer lands the sealed API below and stops
`ApplyChildProcessEnv` from returning an `env`-prefixed argv.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — agent-run / HeadlessRun path that needs color, config-home, PATH
  prepend, and user `-e` without wrapping the child in `env(1)`.
- **Policy inputs** — `runnerID`, `configHome`, `prependPaths`, `envEntries`
  (`KEY=VALUE`), `color`, injectable `parentTERM` (parallel-safe; no process
  `TERM` mutation).
- **`BuildChildProcessEnv`** — pure policy builder → `ChildEnvSpec{Set, Unset}`.
- **`ChildEnvSpec`** — `Set []string` (`KEY=VALUE`) and `Unset []string` (keys)
  for HeadlessRun `CommandEnv` / `CommandUnset`.
- **Pure argv** — remains `codex` / `grok` / …; HeadlessRun gets
  `Command: pureArgv` plus Set/Unset (no `env` binary prefix).

**Behaviors**

- **Empty policy** — no color, no home, no `-e`, no prepend → Set/Unset empty.
- **Color force** — Unset includes `NO_COLOR`; Set has `FORCE_COLOR=1`,
  `CLICOLOR=1`, `CLICOLOR_FORCE=1`; when effective TERM (parent + user `-e`)
  is empty or `dumb` → `TERM=xterm-256color`; otherwise keep effective TERM.
- **Config home** — `codex-tty` → `CODEX_HOME=…`; other runners (e.g. `grok-tty`)
  → `GROK_HOME=…`.
- **PATH prepend** — non-empty `prependPaths` → `PATH=` joined ahead of process
  PATH (same as old `ApplyChildProcessEnv`).
- **User `-e`** — assignments land in Set (order preserved for last-wins at
  merge); with color, `-e NO_COLOR=…` is dropped from Set and covered by Unset.
- **No env(1)** — composition path must not return argv starting with `env`.

**Product API sealed for implementer**

| Symbol | Role |
|--------|------|
| `type ChildEnvSpec struct { Set, Unset []string }` | HeadlessRun CommandEnv / CommandUnset |
| `BuildChildProcessEnv(runnerID, configHome string, prependPaths, envEntries []string, color bool, parentTERM string) ChildEnvSpec` | Pure policy (inject `parentTERM`) |
| `ApplyChildProcessEnv` | Deprecate/remove env prefix: pure argv only (or delete; HeadlessRun uses Build + pure Command) |

HeadlessRun wiring (implementer checklist, not process-spawned here):

```text
spec := BuildChildProcessEnv(runnerID, configHome, prependPaths, envEntries, color, parentTERM)
ttywatch.HeadlessRun(..., Command: pureArgv, CommandEnv: spec.Set, CommandUnset: spec.Unset, ...)
```

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/child-process-env/
├── DOCTEST.md
├── SETUP.md
├── empty-policy/
│   └── no-inputs/                    # S1: Set/Unset empty
├── color/
│   ├── force-keys/
│   │   ├── parent-term-dumb/         # S2: TERM rewrite dumb → xterm-256color
│   │   ├── parent-term-empty/        # S2: empty parent TERM → xterm-256color
│   │   └── parent-term-good/         # S3: good TERM kept (not forced dumb path)
│   └── drops-user-no-color/          # S8: -e NO_COLOR dropped; Unset has NO_COLOR
├── config-home/
│   ├── codex-tty/                    # S4: CODEX_HOME
│   └── grok-tty/                     # S5: GROK_HOME
├── prepend-paths/
│   └── joins-ahead/                  # S6: PATH= prepended
├── user-env/
│   └── sets-foo/                     # S7: FOO=bar in Set
└── pure-argv/
    └── no-env-binary-prefix/         # S9: Apply path argv[0] ≠ "env"
```

Parameter ranking (most → least significant):

1. **Policy surface** — empty / color / config-home / prepend / user-env / pure-argv
2. **Color sub-class** — force-keys (TERM policy) vs drops-user-no-color
3. **Concrete inputs** — parent TERM variant; runner id; path/entry values

## Test Index

| # | Leaf | Scenario | Description |
|---|------|----------|-------------|
| 1 | `empty-policy/no-inputs` | S1 | no color/home/-e/prepend → Set & Unset empty |
| 2 | `color/force-keys/parent-term-dumb` | S2 | color: force keys + Unset NO_COLOR + TERM=xterm-256color |
| 3 | `color/force-keys/parent-term-empty` | S2 | color: empty parentTERM → TERM=xterm-256color |
| 4 | `color/force-keys/parent-term-good` | S3 | color: parentTERM=xterm kept (not dumb rewrite) |
| 5 | `color/drops-user-no-color` | S8 | color + `-e NO_COLOR=1` → NO_COLOR not in Set; in Unset |
| 6 | `config-home/codex-tty` | S4 | configHome + codex-tty → CODEX_HOME=… |
| 7 | `config-home/grok-tty` | S5 | configHome + grok-tty → GROK_HOME=… |
| 8 | `prepend-paths/joins-ahead` | S6 | prependPaths → PATH starts with joined prefixes |
| 9 | `user-env/sets-foo` | S7 | `-e FOO=bar` → FOO=bar in Set |
| 10 | `pure-argv/no-env-binary-prefix` | S9 | ApplyChildProcessEnv (or successor) argv does not start with `env` |

## How to Run

```sh
# from agent-pro module root
doctest vet ./pkgs/agenttty/tests/child-process-env
doctest test ./pkgs/agenttty/tests/child-process-env/...
doctest test -v ./pkgs/agenttty/tests/child-process-env/empty-policy/no-inputs
```

Expect **RED** until implementer exports `BuildChildProcessEnv` / `ChildEnvSpec`
and stops prefixing argv with `env`.

### Planned API

```go
// package agenttty
type ChildEnvSpec struct {
	Set   []string // KEY=VALUE for CommandEnv
	Unset []string // keys for CommandUnset
}

// BuildChildProcessEnv is pure policy. parentTERM is injectable (not os.Getenv).
func BuildChildProcessEnv(runnerID, configHome string, prependPaths, envEntries []string, color bool, parentTERM string) ChildEnvSpec
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/doctest/session"
)

// Request drives pure child-env policy or pure-argv composition checks.
// No t.Setenv / os.Setenv / process environ mutation.
type Request struct {
	// Mode: "build" (default) → BuildChildProcessEnv; "apply-argv" → ApplyChildProcessEnv.
	Mode string

	RunnerID     string
	ConfigHome   string
	PrependPaths []string
	EnvEntries   []string
	Color        bool
	ParentTERM   string // injectable for Build; parallel-safe

	// Argv is the pure agent command for Mode "apply-argv".
	Argv []string
}

// Response holds Set/Unset from Build, or resulting Argv from Apply.
type Response struct {
	Set   []string
	Unset []string
	Argv  []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	mode := req.Mode
	if mode == "" {
		mode = "build"
	}
	switch mode {
	case "apply-argv":
		// Transitional: ApplyChildProcessEnv must stop returning env-prefixed argv.
		// parentTERM is not on the old Apply signature; color/home/-e still exercise prefix.
		out := agenttty.ApplyChildProcessEnv(
			append([]string(nil), req.Argv...),
			req.RunnerID,
			req.ConfigHome,
			req.PrependPaths,
			req.EnvEntries,
			req.Color,
		)
		return &Response{Argv: out}, nil
	case "build":
		spec := agenttty.BuildChildProcessEnv(
			req.RunnerID,
			req.ConfigHome,
			req.PrependPaths,
			req.EnvEntries,
			req.Color,
			req.ParentTERM,
		)
		return &Response{Set: spec.Set, Unset: spec.Unset}, nil
	default:
		t.Fatalf("unknown Mode %q", mode)
		return nil, nil
	}
}
```
