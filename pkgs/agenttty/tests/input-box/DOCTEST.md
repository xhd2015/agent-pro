# Input-box occupancy probe

L2 in-process tests for the exported `agenttty` occupancy probe on snapshot /
scrollback text. Occupancy is **not** `sendable` / `screen_status`.

Live Codex **0.147.0** last-line shapes (see experiment transcript
`~/.sandbox/transcripts/2026-08-14T04:06:55Z-experiment-solution-codex-input-box.md`):

- Empty composer: `›` + placeholder **glued** to ` medium · ` on the **same** line.
- Occupied: `›` + draft only; footer on the **next** line.
- Naive “TrimSpace after `›` is occupied” is **wrong** (placeholder is text).

# DSN (Domain Specific Notion)

Occupancy of the live composer box, from rendered snapshot text.

**Participants**

- **Snapshot text** — tty-watch / `tty snapshot` printable scrollback (ANSI already
  stripped in these fixtures). Injected on `Request.Scrollback` or a testdata file.
- **Composer glyphs** — Codex `›` (U+203A) and `»` (U+00BB); Grok `❯` (U+2771).
- **DetectInputBox** — exported pure probe (name up to implementer; tests call
  `agenttty.DetectInputBox`). Returns `empty` | `occupied` | `unknown`.
- **InputBoxReport** — library status builder mapping a probe token to locked CLI
  presentation (`input box: …` / JSON `input_box`). Shared by `tty status` and
  `status` terminal layer.
- **CheckWritable** — existing sendable classifier. Occupancy must **not** change
  its idle/ready outcome on the live empty and occupied Codex shapes.

**Behaviors**

- Use the **last** composer glyph (live box), not a historical `›`/`»`/`❯` above.
- Codex empty: last `›`/`»` line **contains** ` medium · ` (placeholder glued to
  the model footer) **or** TrimSpace remainder after the glyph is empty.
- Codex occupied: remainder after the last glyph is non-empty **and** that glyph’s
  line does **not** contain ` medium · `.
- Grok: last `❯` line; TrimSpace after glyph empty (padding-only) → `empty`;
  non-empty user text → `occupied`. No Grok footer-glue rule.
- No snapshot, no Codex/Grok glyph → `unknown`. Unreachable TTY / no snapshot
  reports `unknown`.
- Human: `input box: empty|occupied|unknown` (after sendable). JSON field
  `input_box`. CLI last content line ends with `\n`.

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/input-box/
├── DOCTEST.md
├── SETUP.md
├── testdata/                                 # locked 0.147 last-line shapes
├── unknown/                                  # no classifiable composer
│   ├── empty-snapshot/
│   └── no-composer-glyph/
├── codex/                                    # last live glyph is › or »
│   ├── empty/
│   │   ├── live-placeholder-glued/           # 0.147 empty (hint glued to footer)
│   │   ├── remainder-whitespace/             # › + padding only
│   │   └── double-angle-glued/               # » + medium · on same line
│   ├── occupied/
│   │   ├── live-single-line/                 # 0.147 EXP-DRAFT-NOTE-42
│   │   ├── live-multiline/                   # 0.147 LINE1 / LINE2
│   │   ├── placeholder-without-footer/       # hint text, no glue → occupied
│   │   └── double-angle-draft/               # » + draft; footer next line
│   └── last-glyph-wins/
│       └── historical-draft-then-empty/      # leftover › above empty live line
├── grok/                                     # last live glyph is ❯ (conservative)
│   ├── empty-padding/
│   ├── boxed-empty/                          # │ ❯ … │ right border is chrome
│   ├── build-anything-placeholder/           # idle placeholder Build anything → empty
│   ├── occupied-text/
│   └── footer-glue-ignored/                  # medium · on ❯ line still occupied
├── sendable-independent/                     # occupancy ≠ sendable
│   ├── empty-still-idle/
│   └── occupied-still-idle/
└── report/                                   # library status builder
    ├── human-json-empty/
    └── unreachable-unknown/
```

Parameter ranking (most → least significant):

1. **Composer family / presence** — none (unknown) vs last glyph Codex vs Grok
2. **Occupancy evidence on the last glyph line** — footer glue / trim-empty vs draft
3. **Glyph variant** — Codex `›` vs `»`
4. **History** — last live glyph wins over earlier composer lines
5. **Surface** — probe token vs sendable independence vs report presentation

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `unknown/empty-snapshot` | Zero-length scrollback → `unknown` |
| 2 | `unknown/no-composer-glyph` | Chrome without `›`/`»`/`❯` → `unknown` |
| 3 | `codex/empty/live-placeholder-glued` | Live 0.147 empty: hint glued to ` medium · ` → `empty` |
| 4 | `codex/empty/remainder-whitespace` | `›` + TrimSpace-empty remainder → `empty` |
| 5 | `codex/empty/double-angle-glued` | `»` + ` medium · ` on that line → `empty` |
| 6 | `codex/occupied/live-single-line` | Live 0.147 `› EXP-DRAFT-NOTE-42` → `occupied` |
| 7 | `codex/occupied/live-multiline` | Live 0.147 `› LINE1` / `LINE2`+footer → `occupied` |
| 8 | `codex/occupied/placeholder-without-footer` | `› Summarize recent commits` (no glue) → `occupied` |
| 9 | `codex/occupied/double-angle-draft` | `»` + draft; footer on next line → `occupied` |
| 10 | `codex/last-glyph-wins/historical-draft-then-empty` | Historical `› leftover` then empty glued line → `empty` |
| 11 | `grok/empty-padding` | Last `❯` + padding only → `empty` |
| 12 | `grok/boxed-empty` | Boxed `│ ❯ … │` → `empty` |
| 13 | `grok/occupied-text` | Last `❯ leftover note` → `occupied` |
| 14 | `grok/footer-glue-ignored` | `❯` + text containing ` medium · ` → `occupied` |
| 15 | `sendable-independent/empty-still-idle` | Empty glued snapshot still CheckWritable idle |
| 16 | `sendable-independent/occupied-still-idle` | Occupied draft snapshot still CheckWritable idle |
| 17 | `report/human-json-empty` | `input box: empty` / JSON token `empty` |
| 18 | `report/unreachable-unknown` | No snapshot / unreachable → report `unknown` |

## How to Run

```sh
# from agent-pro module root
doctest vet ./pkgs/agenttty/tests/input-box
doctest test ./pkgs/agenttty/tests/input-box
doctest test -v ./pkgs/agenttty/tests/input-box/codex/empty/live-placeholder-glued
doctest test -v ./pkgs/agenttty/tests/input-box/codex/occupied/live-single-line
doctest test -v ./pkgs/agenttty/tests/input-box/report/unreachable-unknown
```

Classic TDD: these leaves are **RED** until `DetectInputBox` and `InputBoxReport`
are exported. Discovery is unlabeled L2 (no `e2e`).

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

// Request injects snapshot text. Scrollback and Fixture are mutually preferred
// as: Fixture basename under testdata/ wins when set; else inline Scrollback.
// Unreachable models TTY down / snapshot miss (no scrollback available).
type Request struct {
	Scrollback  string
	Fixture     string
	Unreachable bool
	TestdataDir string
	ProviderID  string // CheckWritable provider; default codex-tty
	Family      string // grouping tag (unknown|codex|grok|sendable|report)
}

type Response struct {
	InputBox      string
	HumanLine     string // "input box: empty|occupied|unknown"
	JSONValue     string // input_box field token
	WritableReady bool
	WritableState string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	text, err := resolveScrollback(d, req)
	if err != nil {
		return nil, err
	}
	if req.Unreachable {
		text = ""
	}

	// DetectInputBox is the exported pure probe (string-like empty|occupied|unknown).
	// InputBoxReport is the library status builder for tty status / status.
	status := agenttty.DetectInputBox(text)
	token := strings.TrimSpace(fmt.Sprint(status))

	human, jsonVal := agenttty.InputBoxReport(token)

	resp := &Response{
		InputBox:  token,
		HumanLine: human,
		JSONValue: jsonVal,
	}

	providerID := req.ProviderID
	if providerID == "" {
		providerID = "codex-tty"
	}
	if provider, ok := agenttty.Get(providerID); ok && provider.CheckWritable != nil {
		w := provider.CheckWritable([]byte(text))
		resp.WritableReady = w.Ready
		resp.WritableState = w.State
	}
	return resp, nil
}

func resolveScrollback(d *session.Doctest, req *Request) (string, error) {
	if req.Fixture != "" {
		dir := req.TestdataDir
		if dir == "" {
			dir = filepath.Join(d.DOCTEST_ROOT, "testdata")
		}
		raw, err := os.ReadFile(filepath.Join(dir, req.Fixture))
		if err != nil {
			return "", fmt.Errorf("read fixture %s: %w", req.Fixture, err)
		}
		return string(raw), nil
	}
	return req.Scrollback, nil
}
```
