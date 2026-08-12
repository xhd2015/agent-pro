# agentrunapi — BuildFollowUpCommand `--prompt-file` auto-spill (P2)

Classic TDD doctests for plan phase **P2 of 2**: when building a ForceNew
follow-up via `agentrunapi.BuildFollowUpCommand`, long prompts auto-spill to a
file and the child `agent-run run` line emits **`--prompt-file`** instead of
embedding the body after `--`.

Isolated nested root so parent `wait-driver` / `follow-up-color` / classify
trees stay **GREEN** while these opts/behaviors are missing (this root is
compile-RED and/or assert-RED until implementer lands the fields + spill).

**Out of scope (P2):** live iTerm e2e; evaluation package changes (consumes
library automatically once BuildFollowUp spills); CLI inject beyond P1
`ResolveRunPrompt` / `--prompt-file` load.

**P1 already landed:** `agent-run run` accepts `--prompt-file` via
`ResolveRunPrompt`. P2 is **library `BuildFollowUpCommand` only**.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** builds a ForceNew follow-up line via `BuildFollowUpCommand`
  (Open profile typical for iTerm write-text).
- **`FollowUpOpts.Prompt`** — prompt body. Delivery decides whether it goes
  after `--` (short) or into a spill file (long), after `TrimSpace` for the
  **rune-count threshold**.
- **`FollowUpOpts.PromptFile`** — if non-empty (after trim), emit
  `--prompt-file=<abs>` using that path; **do not** re-write; **do not** put
  `Prompt` after `--` (caller already spilled).
- **`FollowUpOpts.PromptSpillDir`** — when auto-spill is required, write the
  spill file under this dir. **Tests always inject** a per-case dir under
  `d.DOCTEST_CASE` (parallel-safe; no `t.Setenv` / `t.Chdir` / `os.Setenv`).
  Production default when empty (document only; not asserted here): e.g.
  `AGENT_RUN_HOME/open-prompts` or `os.TempDir`.
- **`BuildFollowUpCommand`** — pure-ish: opts → single shell-quoted line;
  may create a spill file under `PromptSpillDir` when auto-spilling.
- **Threshold (locked):** `utf8.RuneCountInString(strings.TrimSpace(Prompt))`
  — **≤ 600** inline after `--`; **> 600** spill + `--prompt-file`.
  Prefer constant `PromptFileSpillMinRunes = 600` (or unexported + export for
  tests only if needed).
- **Flag (locked):** `--prompt-file` (matches P1 CLI). Equals form
  `--prompt-file=PATH` or two-token `--prompt-file PATH` both accepted by
  harness.

**Behaviors**

```
# Short prompt (≤600 runes), Open=true
Prompt="hello", PromptSpillDir=tmp
  → follow-up line contains `--` and shell-quoted hello
  → line does NOT contain --prompt-file
  → no spill file required under PromptSpillDir

# Long prompt (>600 runes), Open=true
Prompt= strings.Repeat("字", 601)  # or ASCII pad to 601 runes
  PromptSpillDir=tmp, SessionID=sid
  → follow-up contains --prompt-file=… or --prompt-file …
  → follow-up does NOT embed the long prompt body after --
  → spill file exists; content == TrimSpace(Prompt)
  → path absolute preferred
  → line much shorter than embedding the body (write-text safety)

# Explicit PromptFile
PromptFile=/given/path.txt (abs), Prompt may be empty or ignored for argv
  → emit --prompt-file=/given/path.txt (or two-token form)
  → do not write spill under PromptSpillDir; do not put Prompt after --

# Boundary
Prompt = 600×'a'  → still inline (≤600); no --prompt-file
Prompt = 601×'a'  → spill + --prompt-file

# Empty prompt Open=true
Prompt=""
  → current behavior (-- with empty body / open without body per existing rules)
  → no --prompt-file
```

**Delivery priority (implementer)**

```
if PromptFile != "" (after trim):
  emit --prompt-file=abs(PromptFile); no argv prompt body
  // prefer: … --open --prompt-file=PATH  without trailing `--` empty
else if utf8.RuneCountInString(TrimSpace(Prompt)) > 600:
  write spill under PromptSpillDir (when set); emit --prompt-file=abs
else:
  existing -- / prompt behavior (Open/Detach always append `--`, prompt)
```

## Version

0.0.2

## Decision Tree

```text
tests/agentrunapi/follow-up-prompt-file/
├── DOCTEST.md
├── SETUP.md
├── short-open/                # ≤600: inline prompt, no --prompt-file
├── long-open-spill/           # >600: spill + --prompt-file; body off line
├── explicit-prompt-file/      # PromptFile set → use given path; no re-write
├── boundary/
│   ├── SETUP.md
│   ├── at-threshold/          # 600 runes → still inline
│   └── over-threshold/        # 601 runes → spill
└── empty-prompt-open/         # Prompt="" → no --prompt-file
```

Parameter ranking (most → least significant):

1. **Delivery mode** — short inline | long auto-spill | explicit PromptFile | empty
2. **Boundary** — at (600) vs over (601) threshold
3. **Independence** — pure `BuildFollowUpCommand` + case-local spill dir under
   `d.DOCTEST_CASE`; no iTerm, no agent-run binary, no `t.Setenv`/`t.Chdir`

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `short-open` | Short open prompt → after `--`; no `--prompt-file`; spill dir empty |
| 2 | `long-open-spill` | 601×`字` → `--prompt-file`; spill body match; no long body on line |
| 3 | `explicit-prompt-file` | `PromptFile` abs path → emit that path; no auto-spill write |
| 4 | `boundary/at-threshold` | 600 runes → still inline; no `--prompt-file` |
| 5 | `boundary/over-threshold` | 601 runes → spill + `--prompt-file` |
| 6 | `empty-prompt-open` | Empty prompt + Open → no `--prompt-file` |

## How to Run

```sh
# from agent-pro module root (external/agent-pro-master-2026-08-11)
GOWORK=off doctest vet ./tests/agentrunapi/follow-up-prompt-file
GOWORK=off doctest test ./tests/agentrunapi/follow-up-prompt-file

GOWORK=off doctest test -v ./tests/agentrunapi/follow-up-prompt-file/short-open
GOWORK=off doctest test -v ./tests/agentrunapi/follow-up-prompt-file/long-open-spill
GOWORK=off doctest test -v ./tests/agentrunapi/follow-up-prompt-file/explicit-prompt-file
GOWORK=off doctest test -v ./tests/agentrunapi/follow-up-prompt-file/boundary/at-threshold
GOWORK=off doctest test -v ./tests/agentrunapi/follow-up-prompt-file/boundary/over-threshold
GOWORK=off doctest test -v ./tests/agentrunapi/follow-up-prompt-file/empty-prompt-open
```

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

Expect **RED** (compile and/or assertion) until implementer adds
`FollowUpOpts.PromptFile`, `FollowUpOpts.PromptSpillDir`, and auto-spill
emission in `BuildFollowUpCommand`.

**Parallel-safe harness:** always set `PromptSpillDir` under `d.DOCTEST_CASE`;
no process stdio reassignment; no `t.Setenv` / `t.Chdir` / `os.Setenv`.

### Planned production surface (implementer; not written by designer)

```go
// package agentrunapi

// PromptFileSpillMinRunes is the exclusive lower bound for auto-spill:
// rune count of TrimSpace(Prompt) must be > this value to spill.
// Locked: 600.
const PromptFileSpillMinRunes = 600

type FollowUpOpts struct {
	// ...existing fields...

	// PromptFile if non-empty: emit --prompt-file=<abs path>; skip Prompt on
	// argv (caller already spilled). Do not re-write the file.
	PromptFile string
	// PromptSpillDir when auto-spill is required: write spill file under this
	// dir (tests inject per-case). Empty → production default (document only).
	PromptSpillDir string
}

// BuildFollowUpCommand prompt delivery (after flags, before final join):
//   if pf := strings.TrimSpace(opts.PromptFile); pf != "" {
//       abs, err := filepath.Abs(pf) // or require abs
//       remainder = append(remainder, "--prompt-file="+abs)
//       // prefer no trailing "--", prompt for Open/Detach when file mode
//   } else if utf8.RuneCountInString(strings.TrimSpace(opts.Prompt)) > PromptFileSpillMinRunes {
//       // write under PromptSpillDir (or production default)
//       // remainder = append(remainder, "--prompt-file="+absSpill)
//   } else {
//       // existing: Open/Detach → "--", prompt; else non-empty → "--", prompt
//   }
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/doctest/session"
)

// Locked fixtures / constants
const (
	// promptFileSpillMinRunes mirrors planned PromptFileSpillMinRunes.
	promptFileSpillMinRunes = 600
	flagPromptFile          = "--prompt-file"
	fixtureShortPrompt      = "hello"
)

// Request drives BuildFollowUpCommand with prompt-file / spill focus.
type Request struct {
	SessionID      string
	Prompt         string
	AgentRunner    string
	Open           bool
	Detach         bool
	PromptFile     string // explicit path (abs preferred)
	PromptSpillDir string // injectable auto-spill dir (tests always set when needed)
}

// Response is the harness observation.
type Response struct {
	FollowUp  string
	ErrString string

	// Extracted after BuildFollowUpCommand (best-effort for asserts).
	PromptFilePath   string // value of --prompt-file from FollowUp line
	HasPromptFile    bool
	SpillDirEntries  []string // basenames under PromptSpillDir after call
	SpillFileContent string   // content of PromptFilePath when readable
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	// Classic RED until FollowUpOpts.PromptFile / PromptSpillDir exist and
	// BuildFollowUpCommand implements auto-spill / explicit file delivery.
	line, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{
		SessionID:      req.SessionID,
		Prompt:         req.Prompt,
		AgentRunner:    req.AgentRunner,
		Open:           req.Open,
		Detach:         req.Detach,
		PromptFile:     req.PromptFile,
		PromptSpillDir: req.PromptSpillDir,
	})
	resp := &Response{FollowUp: line}
	if err != nil {
		resp.ErrString = err.Error()
		return resp, nil
	}

	if path, ok := promptFileFlagValue(line); ok {
		resp.HasPromptFile = true
		resp.PromptFilePath = path
		if path != "" {
			if data, rerr := os.ReadFile(path); rerr == nil {
				resp.SpillFileContent = string(data)
			}
		}
	}

	if dir := strings.TrimSpace(req.PromptSpillDir); dir != "" {
		entries, rerr := os.ReadDir(dir)
		if rerr == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				resp.SpillDirEntries = append(resp.SpillDirEntries, e.Name())
			}
		}
	}
	return resp, nil
}

// ensureSpillDir creates PromptSpillDir under d.DOCTEST_CASE and returns abs path.
// Parallel-safe: case dir is leaf-local; no Chdir/Setenv.
func ensureSpillDir(t *testing.T, d *session.Doctest, name string) (string, error) {
	t.Helper()
	if d == nil || strings.TrimSpace(d.DOCTEST_CASE) == "" {
		return "", fmt.Errorf("DOCTEST_CASE unavailable")
	}
	dir := filepath.Join(d.DOCTEST_CASE, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir, nil
	}
	return abs, nil
}

// writeCaseFile writes content under d.DOCTEST_CASE and returns the absolute path.
func writeCaseFile(t *testing.T, d *session.Doctest, name, content string) (string, error) {
	t.Helper()
	if d == nil || strings.TrimSpace(d.DOCTEST_CASE) == "" {
		return "", fmt.Errorf("DOCTEST_CASE unavailable")
	}
	path := filepath.Join(d.DOCTEST_CASE, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}

func shellTokens(line string) []string {
	return strings.Fields(line)
}

func promptFileFlagValue(line string) (string, bool) {
	toks := shellTokens(line)
	for i := 0; i < len(toks); i++ {
		raw := strings.Trim(toks[i], `"'`)
		if raw == flagPromptFile {
			if i+1 < len(toks) {
				return strings.Trim(toks[i+1], `"'`), true
			}
			return "", true
		}
		if strings.HasPrefix(raw, flagPromptFile+"=") {
			return strings.TrimPrefix(raw, flagPromptFile+"="), true
		}
	}
	return "", false
}

func hasPromptFileFlag(line string) bool {
	_, ok := promptFileFlagValue(line)
	return ok
}

// promptBodyAfterDashDash returns tokens after a standalone "--" separator
// joined by space (empty if none / nothing after).
func promptBodyAfterDashDash(line string) string {
	toks := shellTokens(line)
	for i := 0; i < len(toks); i++ {
		raw := strings.Trim(toks[i], `"'`)
		if raw == "--" {
			if i+1 >= len(toks) {
				return ""
			}
			rest := make([]string, 0, len(toks)-i-1)
			for _, t := range toks[i+1:] {
				rest = append(rest, strings.Trim(t, `"'`))
			}
			return strings.Join(rest, " ")
		}
	}
	return ""
}

func runeCountTrimmed(s string) int {
	return utf8.RuneCountInString(strings.TrimSpace(s))
}

func longCJKPrompt(n int) string {
	return strings.Repeat("字", n)
}

func longASCIIPrompt(n int) string {
	return strings.Repeat("a", n)
}
```
