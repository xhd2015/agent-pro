# agent-run open attach fidelity (mouse / alt-screen)

Regression tree for **interactive host fidelity** when production `--open`
attaches with OpenCloseExits (`attach_mode=open` with OpenCloseExits).

**Stack matches the crime scene:** `grok-tty` + **`llm-mock-run-grok`** + real
`grok` (mock LLM server sibling). Formal asserts encode **desired** product
behavior: host must still receive the child's mouse-tracking modes after
`attach_mode=open` AttachWriter.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --keep-tty --agent-runner grok-tty` under isolated
  `AGENT_RUN_HOME`.
- **llm-mock-run-grok** — built from `./agent/llm/llm-mock/llm-mock-run-grok`,
  passed as `--agent-runner-binary`. Isolates `GROK_HOME` via
  `--agent-runner-config-home`. Sibling **`llm-mock`** next to the binary.
- **real grok** — default child of llm-mock-run-grok when
  `LLM_MOCK_RUN_GROK_COMMAND` and `AGENT_RUN_GROK_TTY_COMMAND` are unset.
- **ptywrap + ttywatch.AttachWriter** — production open path uses
  `attach_mode=open` (OpenCloseExits); this tree attaches with `"open"`
  explicitly and captures host-visible bytes under a PTY.

**Behaviors (desired)**

```
agent-run run --keep-tty --agent-runner grok-tty \
  --agent-runner-binary llm-mock-run-grok \
  --agent-runner-config-home <GROK_HOME> \
  --dir <ws> "probe"
  -> real Grok enables mouse + alt-screen
  -> AttachWriter(attach_mode=open) under host PTY
  -> host bytes MUST contain mouse CSI (?1002h and/or ?1006h)
  -> host bytes SHOULD contain alt-screen (?1049h)
```

Fixed: attach_mode=open = writer lifecycle + raw scrollback first frame.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/open-attach-fidelity/
├── DOCTEST.md
├── SETUP.md
├── doc_test.go
└── host-mouse-modes/
    └── screen-attach-preserves-child-mouse/   # label e2e; skip if no grok
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `host-mouse-modes/screen-attach-preserves-child-mouse` | llm-mock-run-grok + real grok; host `screen` preserves mouse CSI |

## How to Run

```sh
# Requires real `grok` on PATH (llm-mock-run-grok default child).
doctest test --label e2e -v ./cmd/agent-run/tests/open-attach-fidelity/host-mouse-modes/screen-attach-preserves-child-mouse
doctest vet ./cmd/agent-run/tests/open-attach-fidelity
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot          string
	BinPath           string // agent-run
	LLMMockRunGrok    string // llm-mock-run-grok
	AttachHelper      string
	Home              string // AGENT_RUN_HOME
	GrokHome          string // --agent-runner-config-home
	Workspace         string
	SessionID         string
	HostAttachMode    string
	ListenAddr        string
	TerminalSessionID string
}

type Response struct {
	HostBytes          []byte
	ControlAttachBytes []byte
	Stderr             string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runOpenAttachFidelity(t, d, req)
}
```
