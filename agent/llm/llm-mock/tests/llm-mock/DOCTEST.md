# llm-mock Tests

Doc-style tests for the OpenAI-compatible HTTP mock server (`agent/llm/llm-mock`).
The server loads a JSON config of request→response exchanges and returns
pre-configured responses, enabling agent testing without real LLM calls.

# DSN (Domain Specific Notion)

**Participants**

- **llm-mock HTTP server** — OpenAI-compatible listener; loads JSON config (`port`,
  `exchanges[]`) and serves `/v1/chat/completions`, `/v1/models`, and `/admin/requests`.
- **Config loader** — reads `--config`, `LLM_MOCK_CONFIG`, or `LLM_MOCK_CONFIG_FILE`;
  when none provided uses default `{"exchanges": []}`; optionally merges
  `LLM_MOCK_EVENTS_FILE` input exchanges after config exchanges.
- **Exchange matcher** — matches incoming chat requests by index, role, and content
  (exact, substring, or empty=any); sequential counter for `index=-1`.
- **Preset catalog** — `--mock-events-preset=list` prints named preset catalog to stdout
  and exits 0 without starting the HTTP server.
- **Preset queue (`genQueue`)** — `--mock-events-preset=<name>` appends resolved
  `AgentEvent` sequences after config load; global FIFO dequeue per HTTP serve (Option A).
- **Random generator** — after prefix exchanges and preset queue are consumed, dequeues
  `AgentEvent`s from `events.NewMockEventStream(seed, prompt)` and converts to OpenAI
  chat completion shape.
- **HTTP clients** — test harness, agent binaries (opencode, pi), or orchestrators
  that POST chat completion requests and read JSON/SSE responses.
- **Admin recorder** — logs each received request for inspection via `/admin/requests`
  or `--events-file` JSONL output.
- **HTTP debug logger** — optional `--log-http <file>.jsonl` writes one JSONL record per
  HTTP exchange (full request/response headers and bodies, or streaming `chunks[]`).

**Behaviors**

- Server announces listening port on stdout (e.g. `:8080`) at startup.
- Non-streaming responses return configured content, tool_calls, generated events after
  prefix exhaustion, or HTTP 400 when prefix exchange does not match request.
- Streaming mode emits SSE chunks ending with `[DONE]`.
- Port fallback binds next available port when configured port is occupied.
- Events input JSONL appends exchanges after config; duplicate explicit indices error.
- Integration tests spawn real agent binaries pointed at the mock as LLM backend.
- `--mock-events-preset` seeds `genQueue` after config exchanges; unknown name errors at
  startup; `list` is catalog-only (no listener).
- `--log-http` path must end with `.jsonl`; invalid suffix errors at server startup.
- Non-streaming log records use `response.stream=false` and `response.body`; streaming uses
  `response.stream=true` and `response.chunks[]` with raw SSE `data:` lines.

## Version

0.0.2

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
├── port-fallback/                    ephemeral port blocked → server finds next available port
├── config-via-env/                   config via LLM_MOCK_CONFIG env var
├── random-fallback/                  prefix exhausted → GenerateEvents fallback
│   ├── no-prefix-first-request/      empty exchanges, 1 HTTP request → 200 (not 400)
│   ├── prefix-then-random/           1 prefix exchange, 2nd request → 200 generated
│   ├── think-response/               generated ActionThink → non-empty content, stop
│   ├── responds-within-deadline/     repo cwd: first random-fallback response within 3s
│   └── second-turn-multi-message/    turn 1 exhausts stream; turn 2 multi-message → 200 (not no_match)
├── events-input/                     LLM_MOCK_EVENTS_FILE input merge
│   ├── appends-exchanges/            config 1 exchange + events file 1 more → 2nd response
│   └── events-only-no-config/        events file only, no config → works
├── admin-requests/                   GET /admin/requests returns recorded requests
├── agent-events/                     --agent-events-file served AgentEvent JSONL
│   ├── no-duplicate-message-two-requests/ # grok think+message split: 2 HTTP → exactly 1 message logged
│   └── topic-hello-from-user-query-wrapper/ # <user_query>\\nHello\\n</user_query> → topic Hello not tag
├── log-http/                         --log-http full HTTP round-trip JSONL
│   ├── non-stream-round-trip/        POST chat-completions → 1 line, stream=false, request+response.body
│   ├── streaming-chunks/             stream=true → response.stream=true, chunks[] with data: lines
│   └── requires-jsonl-suffix/        --log-http bad.log → startup error, no log file
├── mock-events-preset/               --mock-events-preset genQueue seeding
│   ├── list-catalog/                 --mock-events-preset=list → stdout catalog, exit 0, no server
│   ├── unknown-preset/               nonexistent name → startup error before listen
│   ├── think-message-two-requests/   preset think-message; 2 HTTP → think then message (agent-events)
│   ├── prefix-then-preset/           1 prefix exchange + preset simple → 2nd HTTP preset message
│   ├── preset-then-genstream/        preset simple drained → 2nd HTTP genStream think fallback
│   ├── tool-bash-response/           preset tool-bash → chat tool_calls bash, finish_reason tool_calls
│   └── tool-bash-responses-stream/   preset tool-bash → /v1/responses SSE function_call bash (codex wire)
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
| 14 | `port-fallback` | Ephemeral port blocked → server listens on blocked port+1 |
| 15 | `config-via-env` | Config loaded from LLM_MOCK_CONFIG env var instead of --config flag |
| 16 | `random-fallback/no-prefix-first-request` | Empty `exchanges[]`; first HTTP request returns 200 via generator (not HTTP 400 `no_match`) |
| 17 | `random-fallback/prefix-then-random` | One prefix exchange consumed; second request returns 200 generated response (not replay/400) |
| 18 | `random-fallback/think-response` | First generated `ActionThink` maps to non-empty `message.content` with `finish_reason: stop` |
| 19 | `random-fallback/responds-within-deadline` | Empty `exchanges[]` with server cwd = repo root; first HTTP request returns HTTP 200 within 3s (not blocked on probe execution) |
| 20 | `random-fallback/second-turn-multi-message` | Empty `exchanges[]`; turn 1 consumes think+message; second user turn with history → HTTP 200 (not `no_match`) |
| 21 | `events-input/appends-exchanges` | Config 1 exchange + `LLM_MOCK_EVENTS_FILE` 1 more; second HTTP response from events file |
| 22 | `events-input/events-only-no-config` | Only `LLM_MOCK_EVENTS_FILE`, no config file/env → prefix response works |
| 23 | `admin-requests` | GET /admin/requests returns all recorded requests with correct fields |
| 24 | `agent-events/no-duplicate-message-two-requests` | Empty prefix; 2 HTTP requests (think then message split) → agent-events has think+message only (not 3 lines / duplicate message) |
| 25 | `agent-events/topic-hello-from-user-query-wrapper` | Grok-style `<user_query>\\nHello\\n</user_query>` prompt → generated text uses Hello not `<user_query>` |
| 26 | `log-http/non-stream-round-trip` | `--log-http`; POST chat-completions → 1 JSONL line with request path, `response.stream=false`, status 200 body |
| 27 | `log-http/streaming-chunks` | Streaming request → log has `response.stream=true`, non-empty `chunks[]` including `data:` SSE lines |
| 28 | `log-http/requires-jsonl-suffix` | `--log-http /tmp/http.log` → non-zero exit, `.jsonl` in error, log file not created |
| 29 | `integration/opencode` | Spawns real opencode binary with mock backend; verifies exit 0, output contains "Paris", admin records gpt-4 request |
| 30 | `integration/pi` | Spawns real pi binary with mock backend; verifies exit 0, output contains "Paris", admin records gpt-4 request |
| 31 | `mock-events-preset/list-catalog` | `--mock-events-preset=list` exits 0; stdout contains all MVP preset names; no server |
| 32 | `mock-events-preset/unknown-preset` | `--mock-events-preset=nonexistent` → startup error before listen |
| 33 | `mock-events-preset/think-message-two-requests` | Preset `think-message`, empty exchanges; 2 HTTP → think then message via agent-events |
| 34 | `mock-events-preset/prefix-then-preset` | 1 config exchange + preset `simple`; 2nd HTTP gets preset message (not genStream) |
| 35 | `mock-events-preset/preset-then-genstream` | Preset `simple` (1 message); 2nd HTTP gets genStream think (not no_match) |
| 36 | `mock-events-preset/tool-bash-response` | Preset `tool-bash`; 1 chat completion HTTP → `tool_calls` with bash, `finish_reason: tool_calls` |
| 37 | `mock-events-preset/tool-bash-responses-stream` | Preset `tool-bash`; 1 `/v1/responses` stream → SSE `function_call` bash with `preset-bash` args |

## Coverage

- **Endpoints**: POST /v1/chat/completions (streaming & non-streaming), GET /v1/models, GET /admin/requests
- **Matching**: index>=0, index=-1 (sequential), role match (exact + empty=any), content match (exact, substring, empty=any)
- **Multi-turn**: sequential (-1) and indexed (>=0)
- **Response types**: text content, tool_calls, null content
- **Finish reasons**: stop, tool_calls, length
- **Error**: prefix non-match HTTP 400 (both streaming and non-streaming)
- **Random fallback**: empty prefix, prefix-then-generated, ActionThink content mapping
- **Config loading**: optional default empty config, --config flag, LLM_MOCK_CONFIG env var, LLM_MOCK_EVENTS_FILE input merge (events-only)
- **Server behavior**: port fallback, model echoing, UUID id, usage fields
- **Admin endpoint**: request recording with index, method, path, body
- **Log HTTP**: `--log-http` full round-trip JSONL (non-stream body vs streaming chunks), `.jsonl` suffix validation
- **Mock events preset**: `--mock-events-preset` catalog (`list`), unknown name startup error, genQueue FIFO after prefix, genStream fallback after drain, tool-bash tool_calls
- **Integration**: Real binary spawning (opencode, pi) with mock as backend; verifies client output, exit code, and server-side request recording via /admin/requests

## How to Run

```sh
# Run all llm-mock tests
doctest test ./agent/llm/llm-mock/tests/llm-mock

# Run a specific leaf
doctest test ./agent/llm/llm-mock/tests/llm-mock/chat-completions/basic-exact-match

# Run new admin + integration tests
doctest test ./agent/llm/llm-mock/tests/llm-mock/admin-requests

# Run random-fallback tests
doctest test ./agent/llm/llm-mock/tests/llm-mock/random-fallback/...

# Run events-input tests (including events-only-no-config)
doctest test ./agent/llm/llm-mock/tests/llm-mock/events-input/...

# Run log-http tests
doctest test ./agent/llm/llm-mock/tests/llm-mock/log-http/...

# Run mock-events-preset tests
doctest test ./agent/llm/llm-mock/tests/llm-mock/mock-events-preset/...

# Run integration tests (requires opencode and pi binaries in PATH)
doctest test ./agent/llm/llm-mock/tests/llm-mock/integration/...
```

```go
import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
    "testing"
    "time"
)


type HTTPResponse struct {
    StatusCode int
    Body       string
    Headers    map[string]string
}

type Request struct {
    RepoRoot     string
    BinaryPath   string
    ConfigJSON   string
    Port         int
    BlockPort    int
    // BlockListener, when set, is held open during Run (Setup acquired); Run must not re-bind BlockPort.
    BlockListener net.Listener
    // ExpectedFallbackPort, when >0, ASSERT checks resp.Port == this; else legacy hardcoded checks.
    ExpectedFallbackPort int
    Requests     []string
    Endpoint     string
    Method       string
    UseEnvConfig bool
    // BinaryCmd is the command+args to spawn for binary integration tests.
    // When set, Run executes the external binary instead of making HTTP requests directly.
    BinaryCmd []string
    // BinaryEnv holds extra environment variables for the spawned binary.
    BinaryEnv map[string]string
    // EventsFile is the path to a JSON-lines events file written by the mock server
    // when started with --events-file. runBinary reads this instead of querying /admin/requests.
    EventsFile string
    // AgentEventsFile is the path for --agent-events-file (served AgentEvent JSONL).
    // When empty, Run uses a temp file and populates Response.AgentEventsLines.
    AgentEventsFile string
    // LogHTTPFile is the path for --log-http (full HTTP exchange JSONL).
    // When empty, Run omits the flag and does not populate LogHTTPLines.
    LogHTTPFile string
    // ExpectedOutputContains is a string that must appear in the combined stdout+stderr.
    ExpectedOutputContains string
    // ExpectedExitCode is the expected exit code of the binary (default 0).
    ExpectedExitCode int
    // EventsInputJSONL is JSONL content for LLM_MOCK_EVENTS_FILE (input exchanges appended after config).
    EventsInputJSONL string
    // ServerDir is the working directory for the mock server process (default: temp dir).
    ServerDir string
    // HTTPTimeout caps each HTTP client request (default: 10s).
    HTTPTimeout time.Duration
    // MockEventsPreset is --mock-events-preset value (preset name or "list").
    MockEventsPreset string
    // CatalogOnly runs llm-mock --mock-events-preset=<value> without starting HTTP server.
    CatalogOnly bool
}

type Response struct {
    Responses          []HTTPResponse
    ExitCode           int
    Stdout             string
    Stderr             string
    Port               int
    AgentEventsContent string
    AgentEventsLines   []string
    LogHTTPContent     string
    LogHTTPLines       []string
    Err                error
}

func Run(t *testing.T, req *Request) (*Response, error) {
    if len(req.BinaryCmd) > 0 {
        return runBinary(t, req)
    }

    resp := &Response{}

    if req.CatalogOnly {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        args := []string{"--mock-events-preset", req.MockEventsPreset}
        cmd := exec.CommandContext(ctx, req.BinaryPath, args...)
        var stdoutBuf, stderrBuf bytes.Buffer
        cmd.Stdout = &stdoutBuf
        cmd.Stderr = &stderrBuf
        runErr := cmd.Run()
        resp.Stdout = stdoutBuf.String()
        resp.Stderr = stderrBuf.String()
        if runErr != nil {
            var exitErr *exec.ExitError
            if errors.As(runErr, &exitErr) {
                resp.ExitCode = exitErr.ExitCode()
            } else {
                resp.Err = runErr
            }
        }
        return resp, nil
    }

    // Write config JSON to temp file when provided; omit --config for optional-config leaves.
    var configPath string
    if req.ConfigJSON != "" {
        configPath = filepath.Join(t.TempDir(), "llm-mock-config.json")
        if err := os.WriteFile(configPath, []byte(req.ConfigJSON), 0644); err != nil {
            return nil, fmt.Errorf("write config: %w", err)
        }
    }

    // If BlockListener is set, use it to force port fallback; else bind BlockPort when >0.
    var blockListener net.Listener
    if req.BlockListener != nil {
        blockListener = req.BlockListener
        defer blockListener.Close()
    } else if req.BlockPort > 0 {
        var err error
        blockListener, err = net.Listen("tcp", fmt.Sprintf(":%d", req.BlockPort))
        if err != nil {
            return nil, fmt.Errorf("block port %d: %w", req.BlockPort, err)
        }
        defer blockListener.Close()
    }

    // Start the server
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    agentEventsFile := req.AgentEventsFile
    if agentEventsFile == "" {
        agentEventsFile = filepath.Join(t.TempDir(), "agent-events.jsonl")
    }
    req.AgentEventsFile = agentEventsFile

    args := []string{}
    if configPath != "" && !req.UseEnvConfig {
        args = append(args, "--config", configPath)
    }
    args = append(args, "--agent-events-file", agentEventsFile)
    if req.LogHTTPFile != "" {
        args = append(args, "--log-http", req.LogHTTPFile)
    }
    if req.MockEventsPreset != "" && req.MockEventsPreset != "list" {
        args = append(args, "--mock-events-preset", req.MockEventsPreset)
    }

    cmd := exec.CommandContext(ctx, req.BinaryPath, args...)
    serverDir := t.TempDir()
    if req.ServerDir != "" {
        serverDir = req.ServerDir
    }
    cmd.Dir = serverDir

    env := os.Environ()
    if req.UseEnvConfig && configPath != "" {
        env = append(env, "LLM_MOCK_CONFIG="+configPath)
    }
    if req.EventsInputJSONL != "" {
        eventsInputPath := filepath.Join(t.TempDir(), "llm-mock-events-input.jsonl")
        if err := os.WriteFile(eventsInputPath, []byte(req.EventsInputJSONL), 0644); err != nil {
            return nil, fmt.Errorf("write events input: %w", err)
        }
        env = envWithOverrides(env, map[string]string{"LLM_MOCK_EVENTS_FILE": eventsInputPath})
    }
    cmd.Env = env

    var stderrBuf bytes.Buffer
    cmd.Stderr = &stderrBuf

    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("stdout pipe: %w", err)
    }

    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("start server: %w", err)
    }

    // Read the port from server stdout
    port, err := readPort(stdoutPipe)
    if err != nil {
        cancel()
        waitErr := cmd.Wait()
        resp.Stderr = stderrBuf.String()
        if waitErr != nil {
            var exitErr *exec.ExitError
            if errors.As(waitErr, &exitErr) {
                resp.ExitCode = exitErr.ExitCode()
            } else {
                resp.Err = waitErr
            }
        }
        if resp.Err == nil {
            resp.Err = fmt.Errorf("read port: %w\nstderr: %s", err, resp.Stderr)
        }
        return resp, nil
    }
    resp.Port = port

    // Make HTTP requests
    httpTimeout := req.HTTPTimeout
    if httpTimeout == 0 {
        httpTimeout = 10 * time.Second
    }
    baseURL := fmt.Sprintf("http://localhost:%d", port)
    for i, bodyJSON := range req.Requests {
        httpResp, err := makeHTTPRequestWithTimeout(baseURL, req.Endpoint, req.Method, bodyJSON, httpTimeout)
        if err != nil {
            cancel()
            cmd.Wait()
            resp.Stderr = stderrBuf.String()
            resp.Err = fmt.Errorf("HTTP request %d: %w", i, err)
            return resp, nil
        }
        resp.Responses = append(resp.Responses, httpResp)
    }

    if data, readErr := os.ReadFile(agentEventsFile); readErr == nil {
        resp.AgentEventsContent = string(data)
        resp.AgentEventsLines = readAgentEventLines(resp.AgentEventsContent)
    }
    if req.LogHTTPFile != "" {
        if data, readErr := os.ReadFile(req.LogHTTPFile); readErr == nil {
            resp.LogHTTPContent = string(data)
            resp.LogHTTPLines = readAgentEventLines(resp.LogHTTPContent)
        }
    }

    // Cancel and wait for server shutdown
    cancel()
    err = cmd.Wait()
    resp.Stderr = stderrBuf.String()

    if err != nil {
        if ctx.Err() != nil {
            return resp, nil
        }
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            resp.ExitCode = exitErr.ExitCode()
            return resp, nil
        }
        resp.Err = err
    }
    return resp, nil
}
```
