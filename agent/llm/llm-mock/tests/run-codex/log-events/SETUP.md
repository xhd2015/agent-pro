# Scenario

**Feature**: `llm-mock run --log-events <file.jsonl> codex` captures standard AgentEvent JSONL

```
# run flag validated (.jsonl suffix) before codex
llm-mock run --log-events session.jsonl codex [codex-args...]
orchestrator -> mock server --agent-events-file session.jsonl

# mock serves response -> append agent/event/types AgentEvent per line
fake/real codex -> curl mock /v1/responses -> log AgentEvent (think/message/tool_call) -> session.jsonl
```

## Preconditions

- `--log-events` is a `llm-mock run` subcommand flag only (not shortcut, not server mode).
- Path must end with `.jsonl`; invalid suffix errors before mock/codex start.
- Output shape is `agent/event/types` `AgentEvent` JSONL (`type`, `text`, `tool`, …).
- Must **not** write HTTP `RecordedRequest` shape (`method`, `path`, `body` top-level keys).
- `lessflags` `StopOnFirstArg()` leaves tokens after `codex` as codex argv unchanged.

## Steps

1. Grouping `Setup` documents log-events AgentEvent contract; leaves set `LogEventsPath`, fake codex profile, and assertions.
2. `Run` passes `--log-events` when `Request.LogEventsPath` is non-empty and reads the file post-run.

## Context

- `Request.LogEventsPath` — when set, `Run` invokes `llm-mock run --log-events <path> codex ...`.
- `Response.LogEventsLines` — JSONL lines read from `LogEventsPath` after orchestrator exit.
- `parseAgentEventMaps` — validates each line has `type` and rejects RecordedRequest keys.
- `installFakeCodexEchoArgv` — fake `codex` on PATH for argv passthrough leaves (hook ignores argv).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.UseShortcut = false
	return nil
}
```