# Scenario

**Feature**: `llm-mock run --log-events <file.jsonl> grok` captures standard AgentEvent JSONL

```
# run flag validated (.jsonl suffix) before grok
llm-mock run --log-events session.jsonl grok [grok-args...]
orchestrator -> mock server --agent-events-file session.jsonl

# mock serves response -> append agent/event/types AgentEvent per line
fake/real grok -> curl mock -> log AgentEvent (think/message/tool_call) -> session.jsonl
```

## Preconditions

- `--log-events` is a `llm-mock run` subcommand flag only (not shortcut, not server mode).
- Path must end with `.jsonl`; invalid suffix errors before mock/grok start.
- Output shape is `agent/event/types` `AgentEvent` JSONL (`type`, `text`, `tool`, …).
- Must **not** write HTTP `RecordedRequest` shape (`method`, `path`, `body` top-level keys).
- Mock server uses `--agent-events-file` (separate from `--events-file` RecordedRequest log).
- `lessflags` `StopOnFirstArg()` leaves tokens after `grok` as grok argv unchanged.

## Steps

1. Grouping `Setup` documents log-events AgentEvent contract; leaves set `LogEventsPath`, fake grok profile, and assertions.
2. `Run` passes `--log-events` when `Request.LogEventsPath` is non-empty and reads the file post-run.

## Context

- `Request.LogEventsPath` — when set, `Run` invokes `llm-mock run --log-events <path> grok ...`.
- `Response.LogEventsLines` — JSONL lines read from `LogEventsPath` after orchestrator exit.
- `parseAgentEventMaps` — validates each line has `type` and rejects RecordedRequest keys.
- `installFakeGrokEchoArgv` — fake `grok` on PATH for argv passthrough leaves (hook ignores argv).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// --log-events is only on llm-mock run subcommand, not the shortcut binary.
	req.UseShortcut = false
	return nil
}
```