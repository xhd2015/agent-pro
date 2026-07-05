# Scenario

**Feature**: `llm-mock run --log-http <file.jsonl> opencode` captures full HTTP round-trip JSONL

```
# run flag validated (.jsonl suffix) before opencode
llm-mock run --log-http http.jsonl opencode [opencode-args...]
orchestrator -> mock server --log-http http.jsonl

# mock records each HTTP exchange (Chat Completions API)
fake/real opencode -> curl /v1/chat/completions -> log request+response -> http.jsonl
```

## Preconditions

- `--log-http` is a `llm-mock run` subcommand flag only (not shortcut, not server mode).
- Path must end with `.jsonl`; invalid suffix errors before mock/opencode start.
- Output shape is one JSONL line per HTTP exchange with `request` and `response` objects.
- Opencode primarily uses `POST /v1/chat/completions` against the mock provider (not `/v1/responses`).
- `lessflags` `StopOnFirstArg()` leaves tokens after `opencode` as opencode argv unchanged.

## Steps

1. Grouping `Setup` documents log-http contract; leaves set `LogHTTPPath`, fake opencode profile, assertions.
2. `Run` passes `--log-http` when `Request.LogHTTPPath` is non-empty and reads the file post-run.

## Context

- `Request.LogHTTPPath` — when set, `Run` invokes `llm-mock run --log-http <path> opencode ...`.
- `Response.LogHTTPLines` — JSONL lines read from `LogHTTPPath` after orchestrator exit.
- `parseHTTPExchangeMaps` — validates each line has `request`/`response` HTTP exchange shape.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.UseShortcut = false
	return nil
}
```