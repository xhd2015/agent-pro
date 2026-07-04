# Scenario

**Feature**: `--log-http <file.jsonl>` captures full HTTP request/response round-trips

```
# server startup validates .jsonl suffix
llm-mock --log-http http.jsonl -> HTTP debug logger

# one JSONL line per exchange
HTTP client -> POST /v1/chat/completions -> mock server
mock server -> append request+response record -> http.jsonl
```

## Preconditions

- `--log-http` is a mock server flag (not orchestrator-only).
- Path must end with `.jsonl`; invalid suffix errors before the server listens.
- Each JSONL line records `index`, `timestamp`, `duration_ms`, `request`, `response`.
- Non-streaming responses use `response.stream=false` and `response.body` (parsed JSON object).
- Streaming responses use `response.stream=true` and `response.chunks[]` (raw SSE lines).
- Distinct from `--events-file` (RecordedRequest) and `--agent-events-file` (AgentEvent).

## Steps

1. Grouping `Setup` documents HTTP exchange log contract and `parseHTTPExchangeMaps` helper.
2. Leaves set `LogHTTPFile`, config, and HTTP requests; `Run` passes `--log-http` when set.
3. Leaf `Assert` validates JSONL shape via `Response.LogHTTPLines`.

## Context

- `Request.LogHTTPFile` — when set, `Run` starts server with `--log-http <path>`.
- `Response.LogHTTPLines` — JSONL lines read from `LogHTTPFile` after HTTP requests.
- `parseHTTPExchangeMaps` — validates each line has `request` and `response` objects.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	return nil
}
```