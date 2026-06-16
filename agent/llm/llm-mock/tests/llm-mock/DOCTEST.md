# llm-mock Tests

Doc-style tests for the OpenAI-compatible HTTP mock server (`agent/llm/llm-mock`).
The server loads a JSON config of request→response exchanges and returns
pre-configured responses, enabling agent testing without real LLM calls.

## Decision Tree

```
llm-mock
├── chat-completions/                 POST /v1/chat/completions
│   ├── basic-exact-match/            index=0, exact role+content match → "Paris"
│   ├── sequential-match/             index=-1, single exchange → "Hello, world!"
│   ├── multi-turn/                   two exchanges, both index=-1 → 1st "Paris", 2nd "you asked..."
│   ├── multi-turn-indexed/           two exchanges, index=0 and index=1 → ordered matches
│   ├── role-match-any/               empty role matches any role → matched
│   ├── content-empty-match/          empty content matches any message → matched
│   ├── content-substring/            substring match on last user message → matched
│   ├── finish-reason-length/         finish_reason "length" in response
│   ├── tool-calls/                   tool_calls in response with null content
│   ├── no-match/                     no exchange matches → HTTP 400
│   ├── streaming-basic/              stream=true: SSE chunks + [DONE]
│   └── streaming-no-match/           stream=true + no match → HTTP 400 (non-SSE)
├── models-endpoint/                  GET /v1/models → static model list
├── port-fallback/                    port occupied → server finds next available port
├── config-via-env/                   config via LLM_MOCK_CONFIG env var
├── admin-requests/                   GET /admin/requests returns recorded requests
└── integration/                      Real binary integration tests
    ├── opencode/                     Spawns opencode with mock as LLM backend
    └── pi/                           Spawns pi with mock as LLM backend
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `chat-completions/basic-exact-match` | Single exchange with index=0, role+content exact match, returns "Paris" with finish_reason "stop" |
| 2 | `chat-completions/sequential-match` | Single exchange with index=-1, matched sequentially, counter increments |
| 3 | `chat-completions/multi-turn` | Two exchanges both index=-1, first matches first exchange, second matches second |
| 4 | `chat-completions/multi-turn-indexed` | Two exchanges with explicit indices 0 and 1, requests match in order |
| 5 | `chat-completions/role-match-any` | Exchange with empty role, matches request with "system" role |
| 6 | `chat-completions/content-empty-match` | Exchange with empty content, matches any message content |
| 7 | `chat-completions/content-substring` | Content substring match on the last user message |
| 8 | `chat-completions/finish-reason-length` | Response with finish_reason "length" |
| 9 | `chat-completions/tool-calls` | Response with tool_calls array, content null, finish_reason "tool_calls" |
| 10 | `chat-completions/no-match` | No exchange matches request → HTTP 400 with error JSON |
| 11 | `chat-completions/streaming-basic` | Stream mode SSE: role init chunk, content chunks (~3 chars), finish chunk, [DONE] |
| 12 | `chat-completions/streaming-no-match` | Stream mode with no match → HTTP 400 (before SSE starts) |
| 13 | `models-endpoint` | GET /v1/models returns mock model list with correct JSON structure |
| 14 | `port-fallback` | Configured port occupied → server listens on port+1 |
| 15 | `config-via-env` | Config loaded from LLM_MOCK_CONFIG env var instead of --config flag |
| 16 | `admin-requests` | GET /admin/requests returns all recorded requests with correct fields |
| 17 | `integration/opencode` | Spawns real opencode binary with mock backend; verifies exit 0, output contains "Paris", admin records gpt-4 request |
| 18 | `integration/pi` | Spawns real pi binary with mock backend; verifies exit 0, output contains "Paris", admin records gpt-4 request |

## Coverage

- **Endpoints**: POST /v1/chat/completions (streaming & non-streaming), GET /v1/models, GET /admin/requests
- **Matching**: index>=0, index=-1 (sequential), role match (exact + empty=any), content match (exact, substring, empty=any)
- **Multi-turn**: sequential (-1) and indexed (>=0)
- **Response types**: text content, tool_calls, null content
- **Finish reasons**: stop, tool_calls, length
- **Error**: no-match HTTP 400 (both streaming and non-streaming)
- **Config loading**: --config flag, LLM_MOCK_CONFIG env var
- **Server behavior**: port fallback, model echoing, UUID id, usage fields
- **Admin endpoint**: request recording with index, method, path, body
- **Integration**: Real binary spawning (opencode, pi) with mock as backend; verifies client output, exit code, and server-side request recording via /admin/requests

## How to Run

```sh
# Run all llm-mock tests
doctest test ./agent/llm/llm-mock/tests/llm-mock

# Run a specific leaf
doctest test ./agent/llm/llm-mock/tests/llm-mock/chat-completions/basic-exact-match

# Run new admin + integration tests
doctest test ./agent/llm/llm-mock/tests/llm-mock/admin-requests

# Run integration tests (requires opencode and pi binaries in PATH)
doctest test ./agent/llm/llm-mock/tests/llm-mock/integration/...
```
