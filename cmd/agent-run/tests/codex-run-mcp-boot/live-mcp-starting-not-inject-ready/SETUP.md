# Scenario

**Bug**: live Codex `Starting MCP servers` (from real MCP config) must not be inject-ready

```
llm-mock-run-codex + 8 hang MCP servers in isolated CODEX_HOME
  -> agent-run run --detach (empty prompt)
  -> tty snapshot contains "Starting MCP servers"
  -> BannerDetected false; CheckWritable loading
```

Crime scene: many MCPs / slow boot, `agent-run run` injects on banner (`codex`+`›`)
while writable is still MCP-starting.
`~/.sandbox/transcripts/2026-08-17T175911+08-00-crime-scene-codex-mcp-slow-inject.md`.

## Preconditions

- Root Setup builds binaries and writes hang-MCP toml.
- Real `codex` on PATH (else skip).

## Steps

1. Inherit root Setup (binaries, isolated homes, hang MCP toml).
2. Run detach + poll for MCP-boot chrome.
