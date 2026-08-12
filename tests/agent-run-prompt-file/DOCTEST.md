# agent-run run --prompt-file (P1)

Classic TDD (plan **P1 of 2**): `agent-run run` accepts **`--prompt-file PATH`**,
loads the UTF-8 file body as the initial prompt (same meaning as a positional
prompt), exclusive with a non-empty positional prompt. Missing/unreadable file
→ error. Empty file → empty prompt (same empty-prompt rules as empty positional
for `--open` / `--detach`).

**RED** until implementer:

1. CLI: `String("--prompt-file", &promptFile)` on `agent-run run` + `runHelp`
   documents `--prompt-file PATH`.
2. Pure core `ResolveRunPrompt(positional, promptFile string) (string, error)`
   used by `runHeadless` after flag parse (replace bare
   `prompt := strings.TrimSpace(strings.Join(remaining, " "))` alone).
3. TrimSpace policy matches current positional: final prompt is
   `strings.TrimSpace(...)` of file body or positional.

**Out of scope (P1):** BuildFollowUpCommand auto-spill (P2); OpenInNewTerminal /
iTerm write-text; evaluation changes; live e2e / real Codex.

Crime scene (context only): long prompts on open follow-up write-text corrupt;
P1 adds the CLI flag so open follow-ups can pass a short `--prompt-file=…`
line. Auto-spill is P2.

# DSN (Domain Specific Notion)

**Participants**

- **`agent-run run` CLI** (`pkgs/agentruncli`, `run_cmd.go`) — accepts optional
  `--prompt-file PATH`. Path is absolute or relative to process cwd (tests use
  absolute paths under `d.DOCTEST_CASE`). Omitted/empty flag → positional only.
- **`ResolveRunPrompt`** — pure L2 core (export for doctests; `runHeadless`
  calls the same logic):
  - if `promptFile` non-empty (after trim) **and** positional non-empty
    (after trim) → error (mutually exclusive)
  - if `promptFile` non-empty → `os.ReadFile`, UTF-8 body, `TrimSpace`, return
  - else → `TrimSpace(positional)`
  - missing/unreadable path → non-nil error
- **Positional prompt** — remaining argv after flags, joined with spaces then
  trimmed (today: `strings.TrimSpace(strings.Join(remaining, " "))`).
- **Empty prompt** — empty file (after trim) yields `""`, same as empty
  positional for later `--open` / `--detach` / “prompt is required” rules
  (those rules are **not** re-tested here; only resolution is).

**Behaviors**

```
# Accept / load (L2 pure)
ResolveRunPrompt("", promptFile) where file body is "  hello\n"
  → prompt == "hello"   # TrimSpace of file body

# Exclusive
ResolveRunPrompt("x", promptFile) with any readable file
  → error (mutually exclusive with non-empty positional)

# Missing file
ResolveRunPrompt("", "/no/such/file")
  → error (path does not exist / cannot read)

# Empty file
ResolveRunPrompt("", empty.txt)  # zero-byte or whitespace-only after trim
  → prompt == ""

# Help / wire (L2 pure source; no process stdio swap)
runHelp documents --prompt-file PATH
pkgs/agentruncli registers String("--prompt-file", ...)
```

**Flag name (locked)**

`--prompt-file` (not `--prompt-from-file`).

## Version

0.0.2

## Decision Tree

```text
tests/agent-run-prompt-file/
├── DOCTEST.md
├── SETUP.md
├── resolve/                                   # pure ResolveRunPrompt (L2)
│   ├── SETUP.md
│   ├── file-body/                             # file body → trimmed prompt
│   ├── exclusive-with-positional/             # file + non-empty positional → error
│   ├── missing-file/                          # missing path → error
│   └── empty-file/                            # empty file → ""
└── cli/                                       # agent-run run surface (L2 help + wire)
    ├── SETUP.md
    ├── help-lists-prompt-file/                # run_cmd.go runHelp documents flag
    └── source-wire-flag/                      # pkgs/agentruncli registers flag
```

Parameter ranking (most → least significant):

1. **Surface** — resolve (pure API) | cli (help + source-wire)
2. **Prompt source** — file body | exclusive both | missing | empty
3. **Independence** — pure string/path I/O under `d.DOCTEST_CASE`; no Handle,
   no agent-run binary, no `t.Setenv` / `t.Chdir` / `os.Setenv`, no process
   stdio reassignment

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `resolve/file-body` | File body `"  hello\n"`, no positional → prompt `"hello"` (TrimSpace) |
| 2 | `resolve/exclusive-with-positional` | `--prompt-file` path + positional `"x"` → error exclusive |
| 3 | `resolve/missing-file` | Non-existent path → error |
| 4 | `resolve/empty-file` | Empty file, no positional → prompt `""` |
| 5 | `cli/help-lists-prompt-file` | `run_cmd.go` help surface documents `--prompt-file` |
| 6 | `cli/source-wire-flag` | `pkgs/agentruncli` sources register/document `--prompt-file` |

## How to Run

```sh
# from agent-pro module root (external/agent-pro-master-2026-08-11)
GOWORK=off doctest vet ./tests/agent-run-prompt-file
GOWORK=off doctest test ./tests/agent-run-prompt-file

GOWORK=off doctest test -v ./tests/agent-run-prompt-file/resolve/file-body
GOWORK=off doctest test -v ./tests/agent-run-prompt-file/resolve/exclusive-with-positional
GOWORK=off doctest test -v ./tests/agent-run-prompt-file/resolve/missing-file
GOWORK=off doctest test -v ./tests/agent-run-prompt-file/resolve/empty-file
GOWORK=off doctest test -v ./tests/agent-run-prompt-file/cli/help-lists-prompt-file
GOWORK=off doctest test -v ./tests/agent-run-prompt-file/cli/source-wire-flag
```

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

Expect **RED** (compile and/or assertion) until implementer lands
`ResolveRunPrompt`, CLI flag registration, and `runHelp` docs.

**Parallel-safe harness:** no process stdio reassignment; no `t.Setenv` /
`t.Chdir` / `os.Setenv`. File fixtures live under `d.DOCTEST_CASE` only.
CLI help is a pure read of `pkgs/agentruncli/run_cmd.go` (holds `runHelp`);
registration is a pure source scan.

### Planned production surface (implementer; not written by designer)

```go
// package agentruncli

// ResolveRunPrompt resolves agent-run run's initial prompt from a positional
// string and optional --prompt-file path. Exported for L2 pure doctests;
// runHeadless must use the same policy after flag parse.
//
// Policy:
//   - if promptFile non-empty (TrimSpace) and positional non-empty (TrimSpace)
//     → error (mutually exclusive)
//   - if promptFile non-empty → read file UTF-8 body, TrimSpace, return
//   - else return TrimSpace(positional)
//   - missing/unreadable path → non-nil error
func ResolveRunPrompt(positionalPrompt, promptFile string) (string, error)
```

```go
// package agentruncli — runHeadless (run_cmd.go)

// CLI:
//   String("--prompt-file", &promptFile)
// help documents: --prompt-file PATH
//
// After Parse:
//   prompt, err := ResolveRunPrompt(strings.Join(remaining, " "), promptFile)
//   // or TrimSpace join first; exclusive uses non-empty after TrimSpace either way
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/doctest/session"
)

// Fixtures (locked)
const (
	fixturePromptBody    = "hello"
	fixturePromptFileRaw = "  hello\n" // file body before TrimSpace
	fixturePositional    = "x"
	flagPromptFile       = "--prompt-file"
)

// Request drives one leaf via Mode.
type Request struct {
	// Mode: resolve | cli_help | source_wire
	Mode string

	// resolve
	Positional string // joined remaining argv (pre- or post-trim OK; API trims)
	PromptFile string // absolute path, or empty when flag omitted

	// source_wire: currently only "cli_flag"
	SourceWireTarget string
}

// Response is the harness observation.
type Response struct {
	// resolve
	Prompt    string
	ErrString string

	// cli_help: pure read of pkgs/agentruncli/run_cmd.go
	Stdout         string
	HelpSourceRead bool

	// source_wire
	ScannedFiles int
	FlagFound    bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch strings.TrimSpace(req.Mode) {
	case "resolve":
		return runResolve(t, req, resp)
	case "cli_help":
		return runCLIHelp(t, d, req, resp)
	case "source_wire":
		return runSourceWire(t, d, req, resp)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func runResolve(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	// Classic RED until ResolveRunPrompt exists and implements the policy.
	prompt, err := agentruncli.ResolveRunPrompt(req.Positional, req.PromptFile)
	resp.Prompt = prompt
	if err != nil {
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
			if strings.Contains(src, flagPromptFile) {
				resp.FlagFound = true
			}
		}
	}
	return resp, nil
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

// writeCaseFile writes content under d.DOCTEST_CASE and returns the absolute path.
// Parallel-safe: case dir is leaf-local; no Chdir/Setenv.
func writeCaseFile(t *testing.T, d *session.Doctest, name, content string) (string, error) {
	t.Helper()
	if d == nil || strings.TrimSpace(d.DOCTEST_CASE) == "" {
		return "", fmt.Errorf("DOCTEST_CASE unavailable")
	}
	if err := os.MkdirAll(d.DOCTEST_CASE, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(d.DOCTEST_CASE, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// missingCasePath returns an absolute path under d.DOCTEST_CASE that must not exist.
func missingCasePath(t *testing.T, d *session.Doctest, name string) (string, error) {
	t.Helper()
	if d == nil || strings.TrimSpace(d.DOCTEST_CASE) == "" {
		return "", fmt.Errorf("DOCTEST_CASE unavailable")
	}
	if err := os.MkdirAll(d.DOCTEST_CASE, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(d.DOCTEST_CASE, name), nil
}
```
