# agentrunapi — BuildFollowUpCommand `--env-file` auto-spill

When building a ForceNew follow-up via `agentrunapi.BuildFollowUpCommand`, any
session env entry whose `KEY=VALUE` rune count is **> 64** causes the **full**
env list to spill to a file; the child line emits **`--env-file`** and **no**
inline `-e` flags.

# DSN (Domain Specific Notion)

**Participants**

- **`FollowUpOpts.Env`** — `KEY=VALUE` entries.
- **`FollowUpOpts.EnvFile`** — explicit path; emit `--env-file=<abs>`; no rewrite; no `-e`.
- **`FollowUpOpts.EnvSpillDir`** — auto-spill dir (tests inject under `d.DOCTEST_CASE`).
- **Threshold (locked):** any entry with `utf8.RuneCountInString(TrimSpace(entry)) > 64`
  (`EnvFileSpillMinRunes = 64`).
- **Spill contents (locked):** **all** env entries in one file when any triggers spill.

**Behaviors**

```
# Short only
Env=["NO_COLOR=1"] → inline -e; no --env-file; spill dir empty

# Any long entry
Env=["NO_COLOR=1", "PATH="+pad] (PATH=… >64 runes)
  → --env-file=<abs under EnvSpillDir>
  → no inline -e; long body absent from line
  → file contains ALL entries (short + long)

# Explicit EnvFile
EnvFile=/given.env (even if Env has long entries)
  → emit that path; no auto-spill write

# Boundary
entry with exactly 64 runes → still inline
entry with 65 runes → spill
```

## Version

0.0.1

## Decision Tree

```text
tests/agentrunapi/follow-up-env-file/
├── DOCTEST.md
├── SETUP.md
├── short-inline/              # all short → -e only
├── long-spill/                # any long → --env-file; all entries in file
├── mixed-short-and-long/      # short+long → one file; no -e
├── explicit-env-file/         # EnvFile set → given path; no auto-spill
└── boundary/
    ├── at-threshold/          # 64 runes → inline
    └── over-threshold/        # 65 runes → spill
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `short-inline` | Short env → `-e`; no `--env-file` |
| 2 | `long-spill` | Long PATH → `--env-file`; all entries in file |
| 3 | `mixed-short-and-long` | Short+long → one file; no inline `-e` |
| 4 | `explicit-env-file` | `EnvFile` path emitted; spill dir empty |
| 5 | `boundary/at-threshold` | 64-rune entry stays inline |
| 6 | `boundary/over-threshold` | 65-rune entry spills |

## How to Run

```sh
GOWORK=off doctest vet ./tests/agentrunapi/follow-up-env-file
GOWORK=off doctest test ./tests/agentrunapi/follow-up-env-file
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

const (
	envFileSpillMinRunes = 64
	flagEnvFile          = "--env-file"
)

type Request struct {
	SessionID   string
	Prompt      string
	AgentRunner string
	Open        bool
	Detach      bool
	Env         []string
	EnvFile     string
	EnvSpillDir string
}

type Response struct {
	FollowUp         string
	ErrString        string
	EnvFilePath      string
	HasEnvFile       bool
	SpillDirEntries  []string
	SpillFileContent string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	line, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{
		SessionID:   req.SessionID,
		Prompt:      req.Prompt,
		AgentRunner: req.AgentRunner,
		Open:        req.Open,
		Detach:      req.Detach,
		Env:         req.Env,
		EnvFile:     req.EnvFile,
		EnvSpillDir: req.EnvSpillDir,
	})
	resp := &Response{FollowUp: line}
	if err != nil {
		resp.ErrString = err.Error()
		return resp, nil
	}
	if path, ok := envFileFlagValue(line); ok {
		resp.HasEnvFile = true
		resp.EnvFilePath = path
		if path != "" {
			if data, rerr := os.ReadFile(path); rerr == nil {
				resp.SpillFileContent = string(data)
			}
		}
	}
	if dir := strings.TrimSpace(req.EnvSpillDir); dir != "" {
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

func envEntryAtRunes(n int) string {
	// "E=" + pad → exactly n runes when n >= 2.
	if n < 2 {
		n = 2
	}
	return "E=" + strings.Repeat("x", n-2)
}

func longPATHEntry() string {
	// Always > 64 runes.
	return "PATH=" + strings.Repeat("p", envFileSpillMinRunes)
}

func envFileFlagValue(line string) (string, bool) {
	toks := strings.Fields(line)
	for i := 0; i < len(toks); i++ {
		raw := strings.Trim(toks[i], `"'`)
		if raw == flagEnvFile {
			if i+1 < len(toks) {
				return strings.Trim(toks[i+1], `"'`), true
			}
			return "", true
		}
		if strings.HasPrefix(raw, flagEnvFile+"=") {
			return strings.TrimPrefix(raw, flagEnvFile+"="), true
		}
	}
	return "", false
}

func hasEnvFileFlag(line string) bool {
	_, ok := envFileFlagValue(line)
	return ok
}

func hasInlineEnvFlag(line string) bool {
	toks := strings.Fields(line)
	for _, tok := range toks {
		raw := strings.Trim(tok, `"'`)
		if raw == "-e" || raw == "--env" || strings.HasPrefix(raw, "--env=") {
			return true
		}
	}
	return false
}

var _ = utf8.RuneCountInString
```
