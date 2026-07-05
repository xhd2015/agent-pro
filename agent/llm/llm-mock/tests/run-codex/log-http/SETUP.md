# Scenario

**Feature**: `llm-mock run --log-http <file.jsonl> codex` captures full HTTP round-trip JSONL

```
# run flag validated (.jsonl suffix) before codex
llm-mock run --log-http http.jsonl codex [codex-args...]
orchestrator -> mock server --log-http http.jsonl

# mock records each HTTP exchange (Responses API)
fake/real codex -> curl /v1/responses -> log request+response -> http.jsonl
```

## Preconditions

- `--log-http` is a `llm-mock run` subcommand flag only (not shortcut, not server mode).
- Path must end with `.jsonl`; invalid suffix errors before mock/codex start.
- Output shape is one JSONL line per HTTP exchange with `request` and `response` objects.
- Codex primarily uses `POST /v1/responses` against the mock provider.
- `lessflags` `StopOnFirstArg()` leaves tokens after `codex` as codex argv unchanged.

## Steps

1. Grouping `Setup` documents log-http contract; leaves set `LogHTTPPath`, fake codex profile, assertions.
2. `Run` passes `--log-http` when `Request.LogHTTPPath` is non-empty and reads the file post-run.

## Context

- `Request.LogHTTPPath` — when set, `Run` invokes `llm-mock run --log-http <path> codex ...`.
- `Response.LogHTTPLines` — JSONL lines read from `LogHTTPPath` after orchestrator exit.
- `parseHTTPExchangeMaps` — validates each line has `request`/`response` HTTP exchange shape.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.UseShortcut = false
	return nil
}
```