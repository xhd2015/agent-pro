# Scenario

**Feature**: `llm-mock run --log-http <file.jsonl> grok` captures full HTTP round-trip JSONL

```
# run flag validated (.jsonl suffix) before grok
llm-mock run --log-http http.jsonl grok [grok-args...]
orchestrator -> mock server --log-http http.jsonl

# mock records each HTTP exchange
fake/real grok -> curl mock -> log request+response -> http.jsonl
```

## Preconditions

- `--log-http` is a `llm-mock run` subcommand flag only (not shortcut, not server mode).
- Path must end with `.jsonl`; invalid suffix errors before mock/grok start.
- Output shape is one JSONL line per HTTP exchange with `request` and `response` objects.
- Must **not** emit AgentEvent shape (`type` top-level) or RecordedRequest-only shape.
- Mock server uses `--log-http` (separate from `--log-events` / `--events-file`).
- `lessflags` `StopOnFirstArg()` leaves tokens after `grok` as grok argv unchanged.

## Steps

1. Grouping `Setup` documents log-http contract; leaves set `LogHTTPPath`, fake grok profile, assertions.
2. `Run` passes `--log-http` when `Request.LogHTTPPath` is non-empty and reads the file post-run.

## Context

- `Request.LogHTTPPath` — when set, `Run` invokes `llm-mock run --log-http <path> grok ...`.
- `Response.LogHTTPLines` — JSONL lines read from `LogHTTPPath` after orchestrator exit.
- `parseHTTPExchangeMaps` — validates each line has `request`/`response` HTTP exchange shape.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// --log-http is only on llm-mock run subcommand, not the shortcut binary.
	req.UseShortcut = false
	return nil
}
```