# agentui continuation prompt tests

Doc-style tests for `github.com/xhd2015/agent-pro/pkgs/agentui` conversation
continuation: building a runner prompt from prior `message` events plus the new
user turn.

Resume-id gate tests live in a **nested** tree
(`resolve-runner-prompt/DOCTEST.md`) so this root keeps exercising only
`BuildContinuationPrompt` and stays GREEN while the new API is RED.

# DSN (Domain Specific Notion)

**Web session store** holds `events.jsonl` with alternating user and assistant
`message` events. **agentui.Run** (and web `startAgentRun`) must not pass only
the latest user text to the runner when the session already has transcript
history **and** no provider-native resume id is available.

**BuildContinuationPrompt** is a pure function that reads prior `types.AgentEvent`
rows (user + assistant messages), formats a human-readable prefix such as
`Previous conversation:`, and appends the new user prompt. The current turn's
user line must not be duplicated in the prefix when it is already represented
in `newPrompt`.

**ResolveRunnerPrompt** (nested tree) is the gate used when
`runner_session_id` is resolved from session meta: non-empty resume id → raw
new prompt only; empty → delegate to `BuildContinuationPrompt`.

```
prior events.jsonl message rows -> BuildContinuationPrompt -> combined prompt string
combined prompt -> runner Ask / codex exec -> assistant reply
```

This root's tests call `BuildContinuationPrompt` directly (no subprocess, no store).

## Version

0.0.2

## Decision Tree

```
pkgs/agentui/tests/continuation/
├── DOCTEST.md
├── SETUP.md
├── no-prior-history/                         split: empty transcript
│   └── prompt-equals-new-only/               no "Previous conversation" wrapper
├── with-prior-history/                       split: non-empty transcript
│   ├── build-prompt-includes-prior-messages/ single turn; prefix contains "hi"
│   ├── excludes-duplicate-current-user/      last event = current prompt → not doubled in prefix
│   └── multi-turn-preserves-order/           three turns; User/Assistant order preserved
└── resolve-runner-prompt/                    nested DOCTEST root (ResolveRunnerPrompt gate)
    ├── with-resume-id-skips-history/
    ├── without-resume-id-uses-history/
    └── empty-prior-with-resume-id/
```

Parameter ranking for **this** root (most → least significant):

1. **Prior history** — empty vs non-empty (changes whether a prefix exists)
2. **Turn count** — single prior turn vs multi-turn
3. **Current user dedup** — whether the latest stored user line matches `newPrompt`

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `no-prior-history/prompt-equals-new-only` | Zero prior events → built prompt is only `newPrompt` |
| 2 | `with-prior-history/build-prompt-includes-prior-messages` | user:hi + assistant:hello → follow-up contains `hi` |
| 3 | `with-prior-history/excludes-duplicate-current-user` | Events end with same user text as `newPrompt` → prefix omits duplicate |
| 4 | `with-prior-history/multi-turn-preserves-order` | Two complete turns → prefix lists first user before second user |

Nested tree index: see `resolve-runner-prompt/DOCTEST.md` (leaves 5–7).

## How to Run

```sh
doctest vet ./pkgs/agentui/tests/continuation
doctest test -v ./pkgs/agentui/tests/continuation
doctest vet ./pkgs/agentui/tests/continuation/resolve-runner-prompt
doctest test -v ./pkgs/agentui/tests/continuation/resolve-runner-prompt
doctest test -v ./pkgs/agentui/tests/continuation/with-prior-history/build-prompt-includes-prior-messages
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
	PriorEvents []types.AgentEvent
	NewPrompt   string
}

type Response struct {
	BuiltPrompt string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	built := agentui.BuildContinuationPrompt(req.PriorEvents, req.NewPrompt)
	return &Response{BuiltPrompt: strings.TrimSpace(built)}, nil
}
```
