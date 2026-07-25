# procresolve P2 — FormatTree + EnrichInfo

Classic TDD doctests for plan phase **P2** library APIs on
`github.com/xhd2015/agent-pro/pkgs/procresolve` (nested root under
`tests/procresolve/` so P1 `ResolveFromPID` leaves stay independent and GREEN).

1. `FormatTree(nodes []ProcNode, opts TreeFormatOptions) string`
2. `Options.EnrichInfo` + optional `LookupGrokInfo` → `Result.GrokTitle` /
   `Result.GrokModel` on grok hits

**CLI** is covered by `./tests/proc-resolve-cli/` (separate tree).

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — CLI or library client that needs a printable process tree and/or
  grok session title/model after resolve.
- **`FormatTree`** — pure formatter. Rebuilds parent→child from `PPID`, prints
  each node as `PID` + space + `Cmd`, with tree-cli connectors. Does not call
  Resolve or Lsof.
- **`ResolveFromPID` + enrich** — same resolve pipeline as P1; when
  `opts.EnrichInfo` is true and Kind is `grok` with a SessionID, calls
  `LookupGrokInfo(home, sessionID)` (injectable; production may default to
  `agent/grok/sessions.Info` via `GrokHome`) and fills `GrokTitle` / `GrokModel`.
- **Injectable LookupGrokInfo** — tests always inject; never touch real `~/.grok`.

**Behaviors**

```
# FormatTree
nodes []ProcNode + TreeFormatOptions{ASCII}
  -> multi-line string:
       ASCII=false: ├──  └──  │
       ASCII=true:  +--  `--  |
  -> each data line: connectors + "PID Cmd"
  -> trailing newline preferred

# Enrich
ResolveFromPID(..., Options{EnrichInfo, LookupGrokInfo, ListProcs, Lsof, GrokHome})
  -> hard grok hit from open files (P1 rules)
  -> if EnrichInfo: LookupGrokInfo -> GrokTitle, GrokModel
  -> if !EnrichInfo: GrokTitle and GrokModel stay ""
```

**Locked types (P2 additions)**

```text
TreeFormatOptions
  ASCII bool

FormatTree(nodes []ProcNode, opts TreeFormatOptions) string

Options (additions)
  EnrichInfo bool
  LookupGrokInfo func(home, sessionID string) (title, model string, err error)

Result (additions)
  GrokTitle string
  GrokModel string
```

## Version

0.0.2

## Decision Tree

```
tests/procresolve/p2/
├── DOCTEST.md
├── SETUP.md
├── format-tree/                         # pure FormatTree
│   ├── SETUP.md
│   ├── unicode/                         # ASCII=false → ├── └── │
│   └── ascii/                           # ASCII=true  → +-- `-- |
└── enrich-info/                         # EnrichInfo gate
    ├── SETUP.md
    ├── grok/                            # EnrichInfo true → title/model set
    └── off/                             # EnrichInfo false → title empty
```

Parameter ranking (most → least significant):

1. **API concern** — FormatTree vs EnrichInfo path
2. **Connector style** (format) — Unicode vs ASCII
3. **Enrich gate** — EnrichInfo on vs off

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `format-tree/unicode` | 3-node chain FormatTree ASCII=false → `├──`, `└──`, `│` |
| 2 | `format-tree/ascii` | Same nodes ASCII=true → `+--`, `` `-- ``, `|`; no box-drawing |
| 3 | `enrich-info/grok` | EnrichInfo true + grok hit + inject LookupGrokInfo → title/model set |
| 4 | `enrich-info/off` | EnrichInfo false + same hit + lookup available → title/model empty |

## How to Run

```sh
doctest vet ./tests/procresolve/p2
doctest test ./tests/procresolve/p2

doctest test -v ./tests/procresolve/p2/format-tree/unicode
doctest test -v ./tests/procresolve/p2/format-tree/ascii
doctest test -v ./tests/procresolve/p2/enrich-info/grok
doctest test -v ./tests/procresolve/p2/enrich-info/off
```

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

// FixtureProc is one process row for resolve/enrich injectors.
type FixtureProc struct {
	PID  int
	PPID int
	Cmd  string
}

// FixtureNode is one classified tree node for FormatTree.
type FixtureNode struct {
	PID  int
	PPID int
	Role string
	Cmd  string
}

// Request is filled root→leaf.
// Mode: "format_tree" | "enrich"
type Request struct {
	Mode string

	// enrich (and shared resolve injectors)
	PID       int
	Procs     []FixtureProc
	OpenFiles map[int][]string
	MaxDepth  int
	GrokHome  string
	CodexHome string

	EnrichInfo   bool
	InjectLookup bool
	LookupTitle  string
	LookupModel  string
	LookupErr    string

	// format_tree
	TreeNodes []FixtureNode
	ASCII     bool
}

// Response observes the active Mode.
type Response struct {
	Result   *procresolve.Result
	TreeText string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	switch req.Mode {
	case "format_tree":
		return runFormatTree(t, req)
	case "enrich":
		return runEnrich(t, req)
	default:
		return nil, fmt.Errorf("unknown Request.Mode %q (want format_tree|enrich)", req.Mode)
	}
}

func runFormatTree(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	nodes := make([]procresolve.ProcNode, 0, len(req.TreeNodes))
	for _, n := range req.TreeNodes {
		nodes = append(nodes, procresolve.ProcNode{
			PID:  n.PID,
			PPID: n.PPID,
			Role: n.Role,
			Cmd:  n.Cmd,
		})
	}
	text := procresolve.FormatTree(nodes, procresolve.TreeFormatOptions{
		ASCII: req.ASCII,
	})
	return &Response{TreeText: text}, nil
}

func runEnrich(t *testing.T, req *Request) (*Response, error) {
	t.Helper()

	procs := make([]procresolve.Proc, 0, len(req.Procs))
	for _, p := range req.Procs {
		procs = append(procs, procresolve.Proc{
			PID:  p.PID,
			PPID: p.PPID,
			Cmd:  p.Cmd,
		})
	}
	snap := procs
	files := req.OpenFiles
	if files == nil {
		files = map[int][]string{}
	}

	opts := procresolve.Options{
		GrokHome:   req.GrokHome,
		CodexHome:  req.CodexHome,
		MaxDepth:   req.MaxDepth,
		EnrichInfo: req.EnrichInfo,
		ListProcs: func() []procresolve.Proc {
			return snap
		},
		Lsof: func(pid int) []string {
			return files[pid]
		},
	}
	if req.InjectLookup {
		title := req.LookupTitle
		model := req.LookupModel
		lookupErr := req.LookupErr
		opts.LookupGrokInfo = func(home, sessionID string) (string, string, error) {
			_ = home
			_ = sessionID
			if lookupErr != "" {
				return "", "", fmt.Errorf("%s", lookupErr)
			}
			return title, model, nil
		}
	}

	result, err := procresolve.ResolveFromPID(req.PID, opts)
	return &Response{Result: result}, err
}
```
