# codex-run-mcp-boot

`agent-run run --detach` with **real Codex** via `llm-mock-run-codex` and
**MCP servers in isolated `CODEX_HOME`**. Inject-ready must stay false while
the live TUI shows `Starting MCP servers`.

Crime scene:
`~/.sandbox/transcripts/2026-08-17T175911+08-00-crime-scene-codex-mcp-slow-inject.md`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run** — `run --detach` (empty prompt: no inject; keep-alive daemon)
- **llm-mock-run-codex** + sibling **llm-mock** — real Codex TUI, mock Responses API
- **`LLM_MOCK_EXTRA_MCP_TOML_FILE`** — appended `[mcp_servers.*]` on generated `config.toml`
- **Hang MCP** — stdio children that never answer initialize (`sleep 600`)
- **Isolated homes** — `AGENT_RUN_HOME`, `CODEX_HOME` / `LLM_MOCK_CODEX_HOME`
- **tty snapshot** — `agent-run tty snapshot` of the live PTY
- **BannerDetected / CheckWritable** — `pkgs/agenttty` classifiers `run` uses before inject

**Desired behavior**

```
run --detach --agent-runner codex-tty --agent-runner-binary llm-mock-run-codex \
  --agent-runner-config-home <codex-home>  + extra MCP toml (N hang servers)
  -> live snapshot contains "Starting MCP servers"
  -> CheckWritable loading
  -> BannerDetected false   # run must not inject yet
```

**Today (bug)**

`BannerDetected` is true on that live frame (`codex` + `›`), so `run` with a
prompt injects during MCP boot.

## Version

0.0.1

## Decision Tree

```
cmd/agent-run/tests/codex-run-mcp-boot/
├── DOCTEST.md
├── SETUP.md
└── live-mcp-starting-not-inject-ready/   # real TUI + MCP config
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `live-mcp-starting-not-inject-ready` | Live `Starting MCP servers` → not inject-ready |

## How to Run

```sh
# skip if codex not on PATH
doctest test --label e2e --label codex \
  ./cmd/agent-run/tests/codex-run-mcp-boot

doctest test -v --label e2e --label codex \
  ./cmd/agent-run/tests/codex-run-mcp-boot/live-mcp-starting-not-inject-ready

# refresh embedded L2 fixture from this live frame
CODEX_MCP_BOOT_DUMP_SNAPSHOT=pkgs/agenttty/testdata/codex-writable/codex-mcp-servers-starting.txt \
  doctest test --label e2e --label codex \
  ./cmd/agent-run/tests/codex-run-mcp-boot/live-mcp-starting-not-inject-ready
```

## Types

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	CodexHome string
	Workspace string

	AgentRun        string
	LLMMock         string
	LLMMockRunCodex string
	MockConfigFile  string
	ExtraMCPFile    string

	SessionID   string
	ExecTimeout time.Duration
	MCPPoll     time.Duration
	// SnapshotDumpPath, when set (or CODEX_MCP_BOOT_DUMP_SNAPSHOT), writes the
	// live MCP-boot tty snapshot for testdata refresh.
	SnapshotDumpPath string
}

type Response struct {
	DetachStderr string
	Snapshot     string
	Writable     agenttty.WritableStatus
	BannerReady  bool
	ConfigTOML   string
	SawMCPBoot   bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runDetachAndCaptureMCPBoot(t, req)
}
```
