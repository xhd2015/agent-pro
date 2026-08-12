# checkCodexWritable Snapshot Fixture Tests

Doc-style tests for `agenttty.checkCodexWritable` (codex-tty provider `CheckWritable`) using
rendered tty-watch snapshot text fixtures under `pkgs/agenttty/testdata/codex-writable/`.

# DSN (Domain Specific Notion)

**Participants**

- **codex TTY session** — interactive OpenAI Codex process inside a tty-watch PTY; scrollback
  accumulates boot chrome, update modals, MCP warnings, model status, and the main chat prompt
  (`›` U+203A legacy, `»` U+00BB Codex 0.146+, status lines, …).
- **tty-watch snapshot** — `SnapshotText` / `tty-watch snapshot` returns rendered printable
  text (ANSI stripped for detection) passed to writable checks.
- **checkCodexWritable** — codex-tty `Provider.CheckWritable` heuristic: classifies scrollback
  into `WritableStatus` (`ready`, `state`, `reason`). Consumers such as `FetchStatus`
  `waitForPrompt` treat `ready=true` && `state=idle` as safe to inject `/status`.
- **DetectScreenStatus** — codex-tty screen classifier (`idle` / `banner` / `unknown` / …);
  main-chat prompt glyphs must not leave status stuck at `unknown`.
- **Update available modal** — full-screen picker (`Update available`, menu options under `›`,
  `Press enter to continue`); **not** the main chat prompt — must not be classified idle.
- **Fixture manifest** — `expectations.jsonl` maps each `codex-*.txt` to expected
  `ready`/`state`/`reason` (reason matched as substring when non-empty).
- **Test harness** — loads fixture bytes, calls `agenttty.Get("codex-tty").CheckWritable`,
  compares to manifest expectations or leaf-specific outcomes.

**Behaviors**

- Empty snapshot → `state=unknown`, `reason=no terminal output`.
- `model: loading` (compact `model:loading`) → `state=loading`, `ready=false`.
- Update modal / update-available picker (or menu under `›` that is not chat) →
  `state=loading` (non-idle), `ready=false` — **must not** be `idle`.
- MCP startup incomplete without main prompt glyph → `loading`; with main chat `›` **or**
  `»` present → `idle`.
- Main chat prompt `›` (U+203A) **or** `»` (U+00BB) visible without busy/loading signals →
  `state=idle`, `ready=true`.
- DetectScreenStatus for idle chat with only `›` or only `»` → not `unknown`
  (`banner` or `idle`, matching ›-only class today).
- Busy: queued follow-up, **live** working / esc-to-interrupt in the active turn → `state=busy`.
- Historical working / esc-to-interrupt **above** a settled bottom main-chat `›`/`»` → still
  `idle` (post-turn); full-scrollback busy match is a false negative for WaitDone/send.

## Version

0.0.4

## Decision Tree

```
pkgs/agenttty/tests/codex-writable/
├── DOCTEST.md
├── SETUP.md
├── fixture-table/
│   ├── SETUP.md
│   └── all-expectations-match/              # F1: glob codex-*.txt + expectations.jsonl
└── regression/
    ├── SETUP.md
    ├── update-modal-not-idle/               # F2: live update picker → not idle
    ├── model-loading-not-idle/              # F3: update banner + model loading → loading
    ├── main-prompt-idle/                    # F4: MCP incomplete + main › → idle (compat GREEN)
    ├── empty-snapshot-unknown/              # F5: empty bytes → unknown
    ├── double-angle-prompt-idle/            # F6: Codex 0.146 » only → idle (RED before fix)
    ├── double-angle-mcp-idle/               # F7: MCP incomplete + » only → idle (RED before fix)
    └── historical-working-bottom-prompt-idle/ # F8: historical • Working + bottom › → idle (RED before fix)
```

Parameter ranking (most → least significant):

1. **Scenario class** — full fixture table vs targeted regression
2. **Writable outcome** — not-idle (loading) vs idle vs unknown
3. **Snapshot source** — live update modal vs model-loading vs main prompt vs empty vs double-angle
4. **Prompt glyph** — legacy `›` (U+203A) vs Codex 0.146+ `»` (U+00BB)
5. **Fixture variant** — which live/synthetic capture within an outcome class

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `fixture-table/all-expectations-match` | Every `codex-*.txt` matches `expectations.jsonl` (F1) |
| 2 | `regression/update-modal-not-idle` | Live update picker → `ready=false`, not `idle` (F2) |
| 3 | `regression/model-loading-not-idle` | Update banner + `model: loading` → `state=loading` (F3) |
| 4 | `regression/main-prompt-idle` | MCP incomplete + main `›` → `ready=true`, `idle` (F4, compat) |
| 5 | `regression/empty-snapshot-unknown` | Empty bytes → `state=unknown` (F5) |
| 6 | `regression/double-angle-prompt-idle` | Usage-limit + main `»` only → idle + screen≠unknown (F6, RED before fix) |
| 7 | `regression/double-angle-mcp-idle` | MCP incomplete + main `»` only → idle (F7, RED before fix) |
| 8 | `regression/historical-working-bottom-prompt-idle` | Historical `• Working` + settled bottom `›` → idle (F8, RED before fix) |

## How to Run

```sh
doctest vet ./pkgs/agenttty/tests/codex-writable
doctest test ./pkgs/agenttty/tests/codex-writable
doctest test -v ./pkgs/agenttty/tests/codex-writable/regression/update-modal-not-idle
doctest test -v ./pkgs/agenttty/tests/codex-writable/regression/double-angle-prompt-idle
doctest test -v ./pkgs/agenttty/tests/codex-writable/regression/double-angle-mcp-idle
doctest test -v ./pkgs/agenttty/tests/codex-writable/regression/historical-working-bottom-prompt-idle
doctest test -v ./pkgs/agenttty/tests/codex-writable/fixture-table/all-expectations-match
```

```go
import (

	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/doctest/session"
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
	FixtureFile    string // single fixture basename under TestdataDir
	RunAllFixtures bool   // F1 table: evaluate every expectations.jsonl row
	TestdataDir    string // default: module testdata/codex-writable
	RepoRoot       string
}

type Response struct {
	Status         agenttty.WritableStatus
	FixtureResults []FixtureResult
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.RepoRoot == "" {
		req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", "..", ".."))
	}
	if req.TestdataDir == "" {
		req.TestdataDir = filepath.Join(req.RepoRoot, "pkgs", "agenttty", "testdata", "codex-writable")
	}

	provider, ok := agenttty.Get("codex-tty")
	if !ok {
		return nil, fmt.Errorf("codex-tty provider not registered")
	}
	check := provider.CheckWritable

	resp := &Response{}

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
```
