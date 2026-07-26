# agentui ResolveRunnerPrompt tests

Nested doc-style tree for `agentui.ResolveRunnerPrompt` — the resume-id gate that
chooses runner inject text: raw new prompt when a provider session id is bound,
otherwise `BuildContinuationPrompt` history wrap.

Self-contained (nested `DOCTEST.md`): does not inherit parent continuation root
`Request`/`Run`. Parent tree remains GREEN on `BuildContinuationPrompt` alone.

# DSN (Domain Specific Notion)

**Session meta** may bind `RunnerSessionID` (provider/grok/codex resume id).
**agentui.Run** resolves that id, then must choose the inject prompt:

| Condition | Runner inject prompt |
|-----------|----------------------|
| `resumeID` non-empty (native resume will restore provider context) | Raw `newPrompt` only — no history prefix |
| `resumeID` empty (e.g. web multi-turn with fake-codex) | `BuildContinuationPrompt(prior, newPrompt)` |

**ResolveRunnerPrompt** is the pure helper under test:

```
resumeID + newPrompt + prior events
  -> ResolveRunnerPrompt
  -> raw newPrompt (bound) | continuation wrap (unbound)
```

Key off **whether provider resume id is bound**, not TTY vs non-TTY.
`BuildContinuationPrompt` itself is unchanged; this tree only gates when it runs.

## Version

0.0.2

## Decision Tree

```
pkgs/agentui/tests/continuation/resolve-runner-prompt/
├── DOCTEST.md
├── SETUP.md
├── with-resume-id-skips-history/     resume set + prior → raw new prompt only
├── without-resume-id-uses-history/   resume empty + prior → continuation wrap
└── empty-prior-with-resume-id/       resume set + empty prior → still raw new prompt
```

Parameter ranking (most → least significant):

1. **Resume id** — non-empty vs empty (largest inject-path change)
2. **Prior history** — present vs empty (edge: bound resume with no local transcript)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `with-resume-id-skips-history` | Non-empty resume id + prior history → inject is trimmed new prompt only (no `Previous conversation`, no `hi`) |
| 2 | `without-resume-id-uses-history` | Empty resume id + prior history → contains `Previous conversation`, prior `hi`, and new prompt |
| 3 | `empty-prior-with-resume-id` | Non-empty resume id + empty prior → still only new prompt |

## How to Run

```sh
doctest vet ./pkgs/agentui/tests/continuation/resolve-runner-prompt
doctest test -v ./pkgs/agentui/tests/continuation/resolve-runner-prompt
doctest test -v ./pkgs/agentui/tests/continuation/resolve-runner-prompt/with-resume-id-skips-history
```

```go
import (

	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	// PriorEvents is local transcript history available for unbound multi-turn wrap.
	PriorEvents []types.AgentEvent
	// NewPrompt is the current user turn text.
	NewPrompt string
	// ResumeID is the provider/runner session id (e.g. grok-sess-abc). Empty means unbound.
	ResumeID string
}

type Response struct {
	BuiltPrompt string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	built := agentui.ResolveRunnerPrompt(req.ResumeID, req.NewPrompt, req.PriorEvents)
	return &Response{BuiltPrompt: strings.TrimSpace(built)}, nil
}
```
