# debug-with-user Dialog Tests

Doc-style tests for `github.com/xhd2015/agent-pro/agents/debug-with-user/dialog`,
covering AppleScript string escaping and osascript output parsing. Ask flow,
CLI integration, and agent-pro registration live in nested DOCTEST roots under
`cli/` and `agent-pro/`.

# DSN (Domain Specific Notion)

The debug-with-user CLI presents macOS human-in-the-loop checkpoints via
`osascript`. Callers supply title, message, and preset options; the CLI always
appends a **Customize** button and a cancel button, then runs a two-step flow
when Customize is chosen.

The **dialog** package is the library core:

- **Escape** turns arbitrary user/caller text into safe AppleScript string
  literals (quotes, backslashes, newlines).
- **ParseOsascriptOutput** reads `osascript` stdout lines such as
  `button returned:…` and `text returned:…` into structured fields.
- **Ask** (tested via CLI nested root) orchestrates the alert/dialog sequence,
  supports dry-run via `DEBUG_WITH_USER_DRY_RUN=1` for CI, and emits single-line
  JSON on stdout.

Exit semantics: 0 = answer JSON, 1 = dismissed/cancel, 2 = error (e.g.
non-macOS without dry-run).

## Version

0.0.2

## Decision Tree

```
agents/debug-with-user/tests/
├── DOCTEST.md                         # dialog package (this root)
├── SETUP.md
├── dialog/
│   ├── escape/
│   │   ├── quotes-and-backslashes/    # " and \ → valid AppleScript literals
│   │   └── newlines/                  # embedded newlines escaped
│   └── parse/
│       ├── button-returned/           # button returned:Label
│       └── text-returned/             # text returned:typed text
├── cli/                               # === nested DOCTEST root ===
│   ├── ask/
│   │   ├── dry-run/
│   │   │   ├── button-affirmed/       # preset + affirm → via=button, affirmed=true
│   │   │   ├── button-not-affirmed/   # preset not affirm → affirmed=false
│   │   │   ├── customize/             # Customize → free_text JSON
│   │   │   └── dismissed/             # cancel → exit 1
│   │   └── non-mac/                   # no dry-run on non-darwin → exit 2
│   └── skill-show/                    # skill show prints embedded SKILL.md
└── agent-pro/                         # === nested DOCTEST root ===
    └── register/                      # agent-pro skill debug-with-user --show works
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `dialog/escape/quotes-and-backslashes` | `"` and `\` in input → escaped for AppleScript strings |
| 2 | `dialog/escape/newlines` | Multiline input → newlines escaped, no raw line breaks in literals |
| 3 | `dialog/parse/button-returned` | `button returned:X` line → parsed button label |
| 4 | `dialog/parse/text-returned` | `text returned:Y` line → parsed free-text answer |
| 5 | `cli/ask/dry-run/button-affirmed` | Dry-run picks affirm option → JSON `via=button`, `affirmed=true` |
| 6 | `cli/ask/dry-run/button-not-affirmed` | Dry-run picks non-affirm preset → `affirmed=false` |
| 7 | `cli/ask/dry-run/customize` | Dry-run Customize + text → `via=free_text`, no `affirmed` |
| 8 | `cli/ask/dry-run/dismissed` | Dry-run cancel → exit 1, no success JSON |
| 9 | `cli/ask/non-mac` | Without dry-run on non-darwin → exit 2, helpful macOS message |
| 10 | `cli/skill-show` | `debug-with-user skill show` contains `name: debug-with-user` |
| 11 | `agent-pro/register` | `agent-pro skill debug-with-user --show` succeeds after registration |

## How to Run

```sh
doctest vet ./agents/debug-with-user/tests
doctest test -v ./agents/debug-with-user/tests
doctest test -v ./agents/debug-with-user/tests/dialog/escape/quotes-and-backslashes
doctest test -v ./agents/debug-with-user/tests/cli/...
doctest test -v ./agents/debug-with-user/tests/agent-pro/register
```

```go
import (

	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agents/debug-with-user/dialog"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Operation string // escape | parse
	Input     string
}

type Response struct {
	Escaped string
	Button  string
	Text    string
	Err     error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Operation {
	case "escape":
		resp.Escaped = dialog.Escape(req.Input)
	case "parse":
		button, text, err := dialog.ParseOsascriptOutput(req.Input)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Button = button
		resp.Text = text
	default:
		return nil, fmt.Errorf("unknown operation %q", req.Operation)
	}
	return resp, nil
}

func assertEscapedContains(t *testing.T, escaped, substr string) {
	t.Helper()
	if !strings.Contains(escaped, substr) {
		t.Fatalf("escaped %q missing %q", escaped, substr)
	}
}

func assertEscapedNoRawNewline(t *testing.T, escaped string) {
	t.Helper()
	if strings.Contains(escaped, "\n") {
		t.Fatalf("escaped literal must not contain raw newline: %q", escaped)
	}
}
```
