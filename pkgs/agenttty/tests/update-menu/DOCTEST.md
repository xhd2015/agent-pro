# agenttty Update Menu Classifiers

Signed-fixture tests for Codex **Update available** menu vs residual banner
classification and writable impact.

Fixtures: `testdata/update-modal-skip/` (copied from ai-critic signed captures;
see `PROTOCOL.md` for SHA-256).

# DSN (Domain Specific Notion)

**Participants**

- **Signed snapshot fixtures** — live captures from codex-cli 0.143.0 under
  `testdata/update-modal-skip/*.snapshot.txt`.
- **Menu classifiers (`agenttty`)** — production helpers under test:
  - `IsBlockingUpdateMenu(text) bool` — true only for the blocking menu modal
  - `UpdateMenuSelection(text) string` — `UPDATE_NOW` | `SKIP` | `SKIP_UNTIL_NEXT` | `""`
  - `checkCodexWritable` / `codex-tty` `CheckWritable` — residual banner alone must
    **not** force permanent `codex update available` loading.
- **Test harness** — loads fixture text (optional strip of `model:loading`),
  calls classifiers + CheckWritable, returns structured fields.

**Behaviors**

- Fixture `01`: blocking menu, selection `UPDATE_NOW`; writable not idle (update).
- Fixture `02`: blocking menu, selection `SKIP`; writable not idle (update).
- Fixture `03b`: residual banner only (`IsBlocking=false`); writable reason must not
  be bare `update available` (may still load for `model:loading`).
- Fixture `04` with model:loading stripped: residual banner alone → idle ready.

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/update-menu/
├── DOCTEST.md
├── SETUP.md
├── testdata/update-modal-skip/          # COPIED signed fixtures
└── classify/
     ├── blocking-menu/
     │    ├── default-update-now/        # 01 → UPDATE_NOW
     │    └── skip-selected/             # 02 → SKIP
     ├── residual-banner/
     │    ├── menu-dismissed/            # 03b → not menu; reason ≠ update available
     │    └── banner-alone-idle/         # 04 stripped → idle
     └── trust-prompt/
          └── not-update-menu/           # 06 trust ≠ update menu; not sendable as update
```

Parameter ranking (most → least significant):

1. **Screen kind** — blocking menu vs residual banner vs directory trust
2. **Selection / banner variant** — UPDATE_NOW vs SKIP; dismissed vs banner-alone idle
3. **Fixture file** — which signed capture

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `classify/blocking-menu/default-update-now` | `01` → blocking + `UPDATE_NOW` + writable loading(update) |
| 2 | `classify/blocking-menu/skip-selected` | `02` → blocking + `SKIP` + writable not idle |
| 3 | `classify/residual-banner/menu-dismissed` | `03b` → not blocking; reason not update-available |
| 4 | `classify/residual-banner/banner-alone-idle` | banner without model:loading → idle ready |
| 5 | `classify/trust-prompt/not-update-menu` | trust modal ≠ update menu; writable not "update available" |

## How to Run

```sh
# from agent-pro module root
doctest vet ./pkgs/agenttty/tests/update-menu
doctest test ./pkgs/agenttty/tests/update-menu/...
doctest test -v ./pkgs/agenttty/tests/update-menu/classify/blocking-menu/default-update-now
```

```go
import (

	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/doctest/session"
)

// Request drives classify leaves for update-menu classifiers.
type Request struct {
	FixtureFile       string // basename under fixtures dir
	StripModelLoading bool   // strip "model: ... loading" for banner-alone idle
	FixturesDir       string
	RepoRoot          string
}

type Response struct {
	SnapshotText   string
	IsBlockingMenu bool
	MenuSelection  string // UPDATE_NOW | SKIP | SKIP_UNTIL_NEXT | ""
	WritableReady  bool
	WritableState  string
	WritableReason string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	if strings.TrimSpace(req.FixtureFile) == "" {
		return nil, fmt.Errorf("FixtureFile required")
	}
	dir := fixturesDir(d, req)
	path := filepath.Join(dir, req.FixtureFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(raw)
	if req.StripModelLoading {
		var b strings.Builder
		for i, line := range strings.Split(text, "\n") {
			if i > 0 {
				b.WriteByte('\n')
			}
			lower := strings.ToLower(line)
			if strings.Contains(lower, "model:") && strings.Contains(lower, "loading") {
				line = strings.ReplaceAll(line, "loading", "gpt-5.5")
				line = strings.ReplaceAll(line, "Loading", "gpt-5.5")
			}
			b.WriteString(line)
		}
		text = b.String()
	}

	resp := &Response{SnapshotText: text}
	resp.IsBlockingMenu = agenttty.IsBlockingUpdateMenu(text)
	resp.MenuSelection = agenttty.UpdateMenuSelection(text)

	provider, ok := agenttty.Get("codex-tty")
	if !ok {
		return nil, fmt.Errorf("codex-tty provider not registered")
	}
	st := provider.CheckWritable([]byte(text))
	resp.WritableReady = st.Ready
	resp.WritableState = st.State
	resp.WritableReason = st.Reason
	return resp, nil
}

func fixturesDir(d *session.Doctest, req *Request) string {
	if req != nil && strings.TrimSpace(req.FixturesDir) != "" {
		return req.FixturesDir
	}
	candidates := []string{
		filepath.Join(d.DOCTEST_ROOT, "testdata", "update-modal-skip"),
		filepath.Join(d.DOCTEST_ROOT, "..", "testdata", "update-modal-skip"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return filepath.Join(d.DOCTEST_ROOT, "testdata", "update-modal-skip")
}
```
