# checkGrokWritable + OpenReady Snapshot Fixture Tests

Doc-style tests for `agenttty.checkGrokWritable` (grok-tty provider `CheckWritable`) and
**open-lifecycle readiness** (`OpenReady` / screen classification) using rendered tty-watch
snapshot text fixtures under `pkgs/agenttty/testdata/grok-writable/`.

# DSN (Domain Specific Notion)

**Participants**

- **grok TTY session** — interactive Grok process inside a tty-watch PTY; scrollback accumulates
  tool output, assistant text, and prompt chrome (`❯`, `Enter:send`, modern `Grok 4.5` chrome, …).
- **tty-watch snapshot** — `SnapshotText` / `tty-watch snapshot` returns rendered printable
  text (ANSI stripped for detection) passed to writable and open-ready checks.
- **checkGrokWritable** — grok-tty `Provider.CheckWritable` heuristic: classifies scrollback
  into `WritableStatus` (`ready`, `state`, `reason`). Send-queue drainer blocks injection
  when `ready=false` (e.g. `state=busy`). **Orthogonal** to open-lifecycle readiness.
- **legacy banner detector** — current-style `bannerDetectedConfig` / exported `BannerDetected`
  for provider `grok` with markers including `GROK_TTY_BANNER` (also matches `Grok ›` and
  substring `grok build`). Used historically by `waitForBannerRemote`.
- **OpenReady** — whether open lifecycle may proceed past banner wait (inject/attach). True for
  legacy banner markers **or** modern starting/busy/idle chrome; **false** for empty boot and
  project-directory modal (modal overrides legacy false-positive on `"grok build"`).
- **screen class** — coarse frame label: `empty` | `starting` | `busy` | `idle` | `modal` |
  (optional `unknown`).
- **Fixture manifest** — `expectations.jsonl` maps each `grok-*.txt` to expected
  `ready`/`state`/`reason` plus optional `banner_detected_legacy` / `open_ready` / `screen_class`.
- **grok-writable-probe** — debug script capturing live snapshots and (after implement)
  exporting fixtures via `-export-fixtures` / `-from`.
- **Test harness** — loads fixture bytes, calls `CheckWritable` (shared `Run`); open-ready
  leaves call exported `OpenReady` / `ClassifyGrokScreen` / `BannerDetected` in `Assert`.

**Behaviors**

- Empty snapshot → writable `state=unknown`, `reason=no terminal output`; `open_ready=false`,
  `screen_class=empty`; legacy banner false.
- Fast-path ready: `response:` or `submitted:` → writable `state=idle`, `ready=true`.
- Prompt visible without busy signals in **prompt region** → writable `state=idle`, `ready=true`.
- `thinking` / active-agent patterns in prompt tail → writable `state=busy`, `ready=false`.
- Full-scrollback substring `working` matching `git working tree status` while prompt idle
  is a **false positive** (session-18 bug); correct behavior is `ready=true`.
- Banner without prompt (`GROK_TTY_BANNER`) → writable `state=loading`; still open-ready via legacy.
- **Modern starting** (`Starting session…`, `❯`, `Grok 4.5`, `Shift+Tab:mode`) → writable option A:
  `ready=true`, `state=idle`; legacy banner **false**; `open_ready=true`, `screen_class=starting`.
- **Modern busy** (Tasks / `Thinking…` with `❯` chrome) → writable `ready=false`, `state=busy`;
  legacy banner **false**; `open_ready=true`, `screen_class=busy`.
- **Modern idle** (post-turn `❯` + `Turn completed`) → writable `ready=true`, `state=idle`;
  legacy banner **false**; `open_ready=true`, `screen_class=idle`.
- Project-directory confirmation modal (`Run Grok Build in a project directory?`,
  `Enter:submit`, radio options) → writable `ready=false`, **not** `idle`; legacy banner
  **true** (FP on `"grok build"`); **`open_ready=false`**, `screen_class=modal`.
- Legacy `Grok ›` / `GROK_TTY_BANNER` frames → legacy banner true; `open_ready=true`.

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/grok-writable/
├── DOCTEST.md
├── SETUP.md
├── fixture-table/
│   ├── SETUP.md
│   └── all-expectations-match/              # F1: glob grok-*.txt + expectations.jsonl (writable only)
├── regression/
│   ├── SETUP.md
│   ├── git-working-tree-idle-prompt/        # F2: session-18 false positive
│   ├── thinking-prompt-tail-busy/           # F3: real busy while thinking
│   ├── empty-snapshot-unknown/              # F4: boot empty → unknown + open_ready false
│   ├── workspace-project-directory-confirm/ # W2: modal not idle; legacy FP; open_ready false
│   ├── modern-starting-session-chrome/      # M1: modern starting open-ready (option A writable)
│   ├── modern-busy-thinking-tasks/          # M2: modern busy open-ready, not sendable
│   ├── modern-idle-input-post-turn/         # M3: modern idle open-ready + sendable
│   └── legacy-angle-banner-open-ready/      # L1: Grok › legacy banner still open-ready
└── probe-export/
    ├── SETUP.md
    └── capture-dir-round-trip/              # F5: -export-fixtures from mini capture
```

Parameter ranking (most → least significant):

1. **Scenario class** — full fixture table vs targeted regression vs probe export
2. **Screen / chrome generation** — empty | modern (starting/busy/idle) | modal | legacy banner | historical regressions
3. **Outcome axis** — writable (ready/state/reason) vs open-lifecycle (banner_legacy / open_ready / screen_class)
4. **Snapshot source** — probe-captured vs SeaTalk modern freeze vs synthesized false-positive
5. **Export path** — existing capture dir → curated testdata (F5)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `fixture-table/all-expectations-match` | Every `grok-*.txt` matches `expectations.jsonl` ready/state/reason (F1; optional open-ready fields ignored) |
| 2 | `regression/git-working-tree-idle-prompt` | Scrollback `git working tree status` + idle prompt → `ready=true` (F2) |
| 3 | `regression/thinking-prompt-tail-busy` | `thinking` in prompt tail → `state=busy` (F3) |
| 4 | `regression/empty-snapshot-unknown` | Empty bytes → writable unknown; open_ready false; screen empty (F4) |
| 5 | `regression/workspace-project-directory-confirm` | Modal → not ready/not idle; banner_legacy true (FP); open_ready false (W2) |
| 6 | `regression/modern-starting-session-chrome` | Modern starting chrome: option A idle/ready; open_ready true; class starting (M1, RED on OpenReady API) |
| 7 | `regression/modern-busy-thinking-tasks` | Modern busy chrome: writable busy; open_ready true; class busy (M2, RED) |
| 8 | `regression/modern-idle-input-post-turn` | Modern idle post-turn: ready idle; open_ready true; class idle (M3, RED) |
| 9 | `regression/legacy-angle-banner-open-ready` | Legacy `Grok ›` → banner_legacy true; open_ready true (L1, compat) |
| 10 | `probe-export/capture-dir-round-trip` | Probe `-export-fixtures` from mini capture produces parseable manifest (F5) |

## How to Run

```sh
doctest vet ./pkgs/agenttty/tests/grok-writable
doctest test ./pkgs/agenttty/tests/grok-writable

# Focus modern open-ready leaves (expect RED until implementer exports OpenReady + classify)
doctest test -v ./pkgs/agenttty/tests/grok-writable/regression/modern-starting-session-chrome
doctest test -v ./pkgs/agenttty/tests/grok-writable/regression/modern-busy-thinking-tasks
doctest test -v ./pkgs/agenttty/tests/grok-writable/regression/modern-idle-input-post-turn
doctest test -v ./pkgs/agenttty/tests/grok-writable/regression/workspace-project-directory-confirm
doctest test -v ./pkgs/agenttty/tests/grok-writable/fixture-table/all-expectations-match

go test ./pkgs/agenttty/ -run TestCheckGrokWritable -v
```

**RED until implementer:** open-ready leaves' `Assert` calls exported `agenttty.OpenReady`,
`agenttty.ClassifyGrokScreen`, and `agenttty.BannerDetected` (see API notes below). Shared `Run`
only exercises `CheckWritable` so F1 writable-only remains buildable. Writable option A for new
modern fixtures should pass F1 once fixtures are present; open-ready asserts fail or fail to
compile until those APIs exist and behave as specified.

**Integration note:** wiring `OpenReady` into `waitForBannerRemote` (so `--open` does not print
`grok TUI banner not detected` on modern chrome) is implementer scope; this tree freezes the
classifier contract only. Fake TUI delayed `GROK_TTY_BANNER` paths must remain green (compat).

### API surface asserted by this tree (exported names)

```go
// OpenReady reports whether grok-tty open lifecycle may proceed past banner wait.
func OpenReady(scrollback []byte) bool

// ClassifyGrokScreen returns coarse class: empty|starting|busy|idle|modal|unknown
func ClassifyGrokScreen(scrollback []byte) string

// BannerDetected wraps legacy bannerDetectedConfig for characterization / fake-TUI markers.
func BannerDetected(scrollback []byte, provider string, markers []string) bool
```

Open-ready regression leaves call these **exported** APIs directly from `ASSERT.md` (not
package-internal test helpers / `TestExported_`). Prefer these stable exports.

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

// Default markers matching grok-tty provider BannerMarkers (legacy detector).
var grokLegacyBannerMarkers = []string{"GROK_TTY_BANNER"}

type FixtureExpectation struct {
	File                 string   `json:"file"`
	Ready                bool     `json:"ready"`
	State                string   `json:"state"`
	Reason               string   `json:"reason,omitempty"`
	Tags                 []string `json:"tags"`
	Source               string   `json:"source,omitempty"`
	BannerDetectedLegacy *bool    `json:"banner_detected_legacy,omitempty"`
	OpenReady            *bool    `json:"open_ready,omitempty"`
	ScreenClass          string   `json:"screen_class,omitempty"`
}

type FixtureResult struct {
	File     string
	Expected FixtureExpectation
	Actual   agenttty.WritableStatus
}

type Request struct {
	FixtureFile    string // single fixture basename under TestdataDir
	RunAllFixtures bool   // F1 table: evaluate every expectations.jsonl row
	TestdataDir    string // default: module testdata/grok-writable
	ExportFromDir  string // F5: probe capture dir with captures.jsonl
	ExportToDir    string // F5: temp export destination
	RepoRoot       string
}

type Response struct {
	Status         agenttty.WritableStatus
	FixtureResults []FixtureResult
	ExportFiles    []string
	ExportManifest []FixtureExpectation
	ProbeOutput    string
	Scrollback     []byte // single-fixture path: raw bytes for open-ready Asserts
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
	resp.Scrollback = text
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
	// Large fixture rows / future wide JSON.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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
