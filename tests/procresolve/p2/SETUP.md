# Scenario

**Feature**: FormatTree connectors and optional Grok title/model enrich (P2 library)

```
# FormatTree
FormatTree(nodes, TreeFormatOptions{ASCII}) -> tree string

# Enrich
ResolveFromPID + EnrichInfo + LookupGrokInfo -> Result.GrokTitle / GrokModel
```

## Preconditions

- Package exports P2 APIs per `p2/DOCTEST.md` DSN (`FormatTree`,
  `TreeFormatOptions`, `Options.EnrichInfo`, `Options.LookupGrokInfo`,
  `Result.GrokTitle`, `Result.GrokModel`).
- No live processes / real lsof / real grok home — inject fixtures only.
- Mode must be set by grouping Setup (`format_tree` or `enrich`).

## Steps

1. Root Setup seeds MaxDepth, homes, empty OpenFiles.
2. Grouping/leaf Setup installs Mode, nodes or resolve fixtures, enrich flags.
3. Run dispatches FormatTree or ResolveFromPID+enrich.
4. Leaf Assert checks TreeText or GrokTitle/GrokModel.

## Context

- Fixture session id (enrich): `019fabcdef-1234-5678-9abc-def012345678`
- Fixture inject title/model: `fixture-grok-title` / `fixture-grok-model`
- GrokHome default: `/tmp/fake-grok-home`

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

const (
	fixtureGrokSessionID = "019fabcdef-1234-5678-9abc-def012345678"
	fixtureGrokTitle     = "fixture-grok-title"
	fixtureGrokModel     = "fixture-grok-model"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.MaxDepth == 0 {
		req.MaxDepth = 16
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	if req.GrokHome == "" {
		req.GrokHome = "/tmp/fake-grok-home"
	}
	if req.CodexHome == "" {
		req.CodexHome = "/tmp/fake-codex-home"
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertResult(t *testing.T, resp *Response) *procresolve.Result {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Result == nil {
		t.Fatal("Result is nil")
	}
	return resp.Result
}

func assertEqualString(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %q, want %q", field, got, want)
	}
}

func assertEqualInt(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d, want %d", field, got, want)
	}
}

func assertContainsString(t *testing.T, field, got, substr string) {
	t.Helper()
	if substr == "" {
		t.Fatalf("%s: empty substr", field)
	}
	if !strings.Contains(got, substr) {
		t.Fatalf("%s: got %q, want substring %q", field, got, substr)
	}
}

func grokSessionPath(uuid string) string {
	return "/tmp/fake-grok-home/.grok/sessions/2026-07/" + uuid + "/events.jsonl"
}
```
