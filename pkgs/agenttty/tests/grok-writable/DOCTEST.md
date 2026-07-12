# checkGrokWritable Snapshot Fixture Tests

Doc-style tests for `agenttty.checkGrokWritable` (grok-tty provider `CheckWritable`) using
rendered tty-watch snapshot text fixtures under `pkgs/agenttty/testdata/grok-writable/`.

# DSN (Domain Specific Notion)

**Participants**

- **grok TTY session** — interactive Grok process inside a tty-watch PTY; scrollback accumulates
  tool output, assistant text, and prompt chrome (`❯`, `Enter:send`, …).
- **tty-watch snapshot** — `SnapshotText` / `tty-watch snapshot` returns rendered printable
  text (ANSI stripped for detection) passed to writable checks.
- **checkGrokWritable** — grok-tty `Provider.CheckWritable` heuristic: classifies scrollback
  into `WritableStatus` (`ready`, `state`, `reason`). Send-queue drainer blocks injection
  when `ready=false` (e.g. `state=busy`).
- **Fixture manifest** — `expectations.jsonl` maps each `grok-*.txt` to expected
  `ready`/`state`/`reason`.
- **grok-writable-probe** — debug script capturing live snapshots and (after implement)
  exporting fixtures via `-export-fixtures` / `-from`.
- **Test harness** — loads fixture bytes, calls `agenttty.Get("grok-tty").CheckWritable`,
  compares to manifest expectations.

**Behaviors**

- Empty snapshot → `state=unknown`, `reason=no terminal output`.
- Fast-path ready: `response:` or `submitted:` → `state=idle`, `ready=true`.
- Prompt visible without busy signals in **prompt region** → `state=idle`, `ready=true`.
- `thinking` / active-agent patterns in prompt tail → `state=busy`, `ready=false`.
- Full-scrollback substring `working` matching `git working tree status` while prompt idle
  is a **false positive** (session-18 bug); correct behavior is `ready=true`.
- Banner without prompt (`GROK_TTY_BANNER`) → `state=loading`.
- Project-directory confirmation modal (`Run Grok Build in a project directory?`,
  `Enter:submit`, radio options) → `ready=false`, **not** `idle` (blocking picker;
  not a session turn prompt).

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/grok-writable/
├── DOCTEST.md
├── SETUP.md
├── fixture-table/
│   ├── SETUP.md
│   └── all-expectations-match/              # F1: glob grok-*.txt + expectations.jsonl
├── regression/
│   ├── SETUP.md
│   ├── git-working-tree-idle-prompt/        # F2: session-18 false positive (RED before fix)
│   ├── thinking-prompt-tail-busy/           # F3: real busy while thinking
│   ├── empty-snapshot-unknown/              # F4: boot empty → unknown
│   └── workspace-project-directory-confirm/ # W2: project-dir picker not idle
└── probe-export/
    ├── SETUP.md
    └── capture-dir-round-trip/              # F5: -export-fixtures from mini capture
```

Parameter ranking (most → least significant):

1. **Scenario class** — full fixture table vs targeted regression vs probe export
2. **Writable outcome** — idle (ready) vs busy vs unknown vs loading
3. **Snapshot source** — probe-captured vs synthesized false-positive vs inline regression
4. **Export path** — existing capture dir → curated testdata (F5)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `fixture-table/all-expectations-match` | Every `grok-*.txt` matches `expectations.jsonl` (F1, 20 fixtures incl. workspace-confirm) |
| 2 | `regression/git-working-tree-idle-prompt` | Scrollback `git working tree status` + idle prompt → `ready=true` (F2, RED before fix) |
| 3 | `regression/thinking-prompt-tail-busy` | `thinking` in prompt tail → `state=busy` (F3) |
| 4 | `regression/empty-snapshot-unknown` | Empty bytes → `state=unknown` (F4) |
| 5 | `regression/workspace-project-directory-confirm` | Project-dir picker → `ready=false`, not idle (W2) |
| 6 | `probe-export/capture-dir-round-trip` | Probe `-export-fixtures` from mini capture produces parseable manifest (F5) |

## How to Run

```sh
doctest vet ./pkgs/agenttty/tests/grok-writable
doctest test ./pkgs/agenttty/tests/grok-writable
doctest test -v ./pkgs/agenttty/tests/grok-writable/regression/git-working-tree-idle-prompt
doctest test -v ./pkgs/agenttty/tests/grok-writable/regression/workspace-project-directory-confirm
doctest test -v ./pkgs/agenttty/tests/grok-writable/fixture-table/all-expectations-match

go test ./pkgs/agenttty/ -run TestCheckGrokWritable -v
```

```go
import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

type FixtureExpectation struct {
	File   string   `json:"file"`
	Ready  bool     `json:"ready"`
	State  string   `json:"state"`
	Reason string   `json:"reason,omitempty"`
	Tags   []string `json:"tags"`
	Source string   `json:"source,omitempty"`
}

type FixtureResult struct {
	File     string
	Expected FixtureExpectation
	Actual   agenttty.WritableStatus
}

type Request struct {
	FixtureFile       string // single fixture basename under TestdataDir
	RunAllFixtures    bool   // F1 table: evaluate every expectations.jsonl row
	TestdataDir       string // default: module testdata/grok-writable
	ExportFromDir     string // F5: probe capture dir with captures.jsonl
	ExportToDir       string // F5: temp export destination
	RepoRoot          string
}

type Response struct {
	Status         agenttty.WritableStatus
	FixtureResults []FixtureResult
	ExportFiles    []string
	ExportManifest []FixtureExpectation
	ProbeOutput    string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.RepoRoot == "" {
		req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", "..", "..", ".."))
	}
	if req.TestdataDir == "" {
		req.TestdataDir = filepath.Join(req.RepoRoot, "pkgs", "agenttty", "testdata", "grok-writable")
	}

	provider, ok := agenttty.Get("grok-tty")
	if !ok {
		return nil, fmt.Errorf("grok-tty provider not registered")
	}
	check := provider.CheckWritable

	resp := &Response{}

	if req.ExportFromDir != "" {
		exportDir := req.ExportToDir
		if exportDir == "" {
			exportDir = filepath.Join(t.TempDir(), "export-out")
		}
		out, err := runProbeExport(t, req.RepoRoot, req.ExportFromDir, exportDir)
		resp.ProbeOutput = out
		if err != nil {
			return resp, err
		}
		manifestPath := filepath.Join(exportDir, "expectations.jsonl")
		manifest, err := loadExpectations(manifestPath)
		if err != nil {
			return resp, fmt.Errorf("load export manifest: %w", err)
		}
		resp.ExportManifest = manifest
		matches, _ := filepath.Glob(filepath.Join(exportDir, "grok-*.txt"))
		resp.ExportFiles = matches
		return resp, nil
	}

	if req.RunAllFixtures {
		manifestPath := filepath.Join(req.TestdataDir, "expectations.jsonl")
		expectations, err := loadExpectations(manifestPath)
		if err != nil {
			return nil, err
		}
		for _, exp := range expectations {
			text, err := os.ReadFile(filepath.Join(req.TestdataDir, exp.File))
			if err != nil {
				return nil, fmt.Errorf("read fixture %s: %w", exp.File, err)
			}
			actual := check(text)
			resp.FixtureResults = append(resp.FixtureResults, FixtureResult{
				File:     exp.File,
				Expected: exp,
				Actual:   actual,
			})
		}
		return resp, nil
	}

	fixture := req.FixtureFile
	if fixture == "" {
		return nil, fmt.Errorf("FixtureFile is required when not running all fixtures")
	}
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixture))
	if err != nil {
		return nil, fmt.Errorf("read fixture %s: %w", fixture, err)
	}
	resp.Status = check(text)
	return resp, nil
}

func loadExpectations(path string) ([]FixtureExpectation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []FixtureExpectation
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var exp FixtureExpectation
		if err := json.Unmarshal([]byte(line), &exp); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		out = append(out, exp)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func runProbeExport(t *testing.T, repoRoot, fromDir, exportDir string) (string, error) {
	t.Helper()
	ctx := exec.Command("go", "run", "./script/debug/grok-writable-probe",
		"-export-fixtures="+exportDir,
		"-from="+fromDir,
	)
	ctx.Dir = repoRoot
	ctx.Env = append(os.Environ(), "HOME="+os.Getenv("HOME"))
	out, err := ctx.CombinedOutput()
	return string(out), err
}
```