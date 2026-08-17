# Scenario

**Bug**: live `Starting MCP servers` chrome must not be inject-ready for `agent-run run`

```
snapshot has "Starting MCP servers (0/8): slowinit_*" + main chat › + OpenAI Codex
  -> CheckWritable is loading (MCP starting)
  -> BannerDetected (waitForBanner / run inject-ready) is false
```

Crime scene: when Codex has many MCPs and boot is slow, `agent-run run` injects as
soon as banner heuristics see `codex` + `›`, while send/status correctly treat the
same frame as not writable. Prompt is lost if the TUI is still in "esc to interrupt"
MCP boot. Transcript:
`~/.sandbox/transcripts/2026-08-17T175911+08-00-crime-scene-codex-mcp-slow-inject.md`.

## Preconditions

- Fixture `codex-mcp-servers-starting.txt` is the live `agent-run tty snapshot`
  from `cmd/agent-run/tests/codex-run-mcp-boot` (`llm-mock-run-codex` + 8 hang MCP
  servers). Refresh:
  `CODEX_MCP_BOOT_DUMP_SNAPSHOT=pkgs/agenttty/testdata/codex-writable/codex-mcp-servers-starting.txt
  doctest test --label e2e --label codex ./cmd/agent-run/tests/codex-run-mcp-boot/live-mcp-starting-not-inject-ready`
- Desired product: `run` hard-waits inject-ready via `BannerDetected` / `bannerDetectedConfig`.
  That must stay **false** until MCP boot chrome is gone (same as `CheckWritable` loading).
- Current product: `codexMCPServersStarting` is checked before `codex` + `›` true-returns.

## Steps

1. Set `req.FixtureFile` to the MCP-servers-starting fixture.

## Context

- F9: inject-ready must agree with writable on live MCP-boot chrome.
- After boot, `MCP startup incomplete` + main `›` remains idle (F4 / F7) — not this fixture.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureMCPServersStarting
	return nil
}
```
