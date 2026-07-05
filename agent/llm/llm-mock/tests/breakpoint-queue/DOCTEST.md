# Breakpoint Queue Tests

Doc-style tests for **breakpoint dequeue** and **API encoders** on the llm-mock HTTP server.
Each HTTP request consumes `genQueue` through exactly one breakpoint (`tool_call` or `message`);
leading `think` events collapse into that breakpoint. Encoders in `agent/llm/openai` and
`agent/llm/anthropic` shape the dequeued slice for `/v1/chat/completions`, `/v1/responses`,
and `/v1/messages`.

# DSN (Domain Specific Notion)

**Participants**

- **llm-mock HTTP server** — serves `/v1/chat/completions`, `/v1/responses`, `/v1/messages`,
  `/v1/models`, and `/admin/requests`; loads JSON config and optional `--mock-events-preset`.
- **Preset queue (`genQueue`)** — FIFO `AgentEvent` sequence seeded by `--mock-events-preset=<name>`
  after config `exchanges[]` prefix is exhausted.
- **Breakpoint dequeuer (`DequeueToBreakpoint`)** — pops leading `think` events, then exactly one
  `tool_call` or `message` breakpoint; never two `tool_call`s in one HTTP response; never consumes
  past the first breakpoint.
- **OpenAI chat encoder** — maps breakpoint slice to Chat Completions JSON (`message.content`,
  `tool_calls`, `finish_reason`); merges prefix think into `message.content` for message breakpoints;
  omits prefix think from chat wire for tool_call breakpoints (logged via agent-events only).
- **OpenAI responses encoder** — maps breakpoint slice to Responses API `output[]` (stream + non-stream);
  emits collapsed think as **reasoning** item before `function_call` (option B); remaps `bash`→`exec_command`,
  `command`→`cmd` in tool arguments (best-effort).
- **Anthropic messages encoder** — maps breakpoint slice to Messages API `content[]` with native
  `{type:"thinking"}` blocks for prefix thinks, then one `{type:"tool_use"}` or `{type:"text"}` breakpoint block.
- **Random generator (`genStream`)** — after `genQueue` drained, existing `events.NewMockEventStream`
  fallback applies (unchanged algorithm).
- **Agent-events logger** — `--agent-events-file` JSONL records every `AgentEvent` consumed per HTTP serve.
- **HTTP clients** — test harness sends sequential POSTs and asserts response bodies + agent-events order.

**Behaviors**

- One breakpoint per HTTP request across all three LLM endpoints (unified semantics; no `drainGenQueue` split).
- `think` is not a breakpoint; multiple leading thinks collapse forward into the next breakpoint.
- `tool_call` breakpoints yield at most one tool per HTTP response (`tool_calls` / `function_call` / `tool_use`).
- After queue empty → `genStream` random fallback (not `no_match`).
- Static `exchanges[]` prefix matching remains unchanged (out of scope for this tree).
- Agent-events log reflects all events consumed on each serve (thinks + breakpoint).

## Version

0.0.2

## Decision Tree

```
breakpoint-queue
├── dequeue/                              breakpoint dequeue semantics (endpoint-agnostic)
│   ├── think-tool-message-two-requests/  preset think-tool-message; 2 chat HTTP → tool then message
│   ├── two-tool-then-message-three-requests/ preset two-tool-message; 3 chat HTTP → tool, tool, message
│   ├── think-merged-into-reply/          preset think-message; 1 chat HTTP → think+message in content
│   ├── think-before-tool-responses-reasoning/ preset think-tool-message; 1 responses stream → reasoning + function_call
│   └── second-request-after-tool/        preset think-tool-message; 2 responses HTTP → message only on #2
├── openai/                               OpenAI wire encoders
│   ├── chat/
│   │   └── tool-bash-single-breakpoint/  preset tool-bash; 1 chat HTTP → one tool_call (regression guard)
│   └── responses/
│       ├── tool-bash-with-reasoning-prefix/ preset think-tool-message; 1 responses stream → reasoning + function_call
│       └── tool-bash-only/               preset tool-bash; 1 responses stream → function_call only
└── anthropic/                            Anthropic Messages API encoder
    └── messages/
        ├── think-tool-message-two-requests/ preset think-tool-message; 2 messages HTTP → thinking+tool_use, then text
        └── tool-bash-one-request/        preset tool-bash; 1 messages HTTP → tool_use block
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `dequeue/think-tool-message-two-requests` | Preset `think-tool-message`; 2 chat HTTP → #1 `tool_calls` (think omitted from wire), #2 message content; agent-events think+tool on #1, message on #2 |
| 2 | `dequeue/two-tool-then-message-three-requests` | Preset `two-tool-message` `[tool_call, tool_call, message]`; 3 chat HTTP → tool, tool, message; one tool per response |
| 3 | `dequeue/think-merged-into-reply` | Preset `think-message`; 1 chat HTTP → single content with think text + message text concatenated |
| 4 | `dequeue/think-before-tool-responses-reasoning` | Preset `think-tool-message`; 1 `/v1/responses` stream → SSE reasoning item with think text + `function_call` (no message on #1) |
| 5 | `dequeue/second-request-after-tool` | Preset `think-tool-message`; 2 `/v1/responses` HTTP → #2 message/`output_text` only |
| 6 | `openai/chat/tool-bash-single-breakpoint` | Preset `tool-bash`; 1 chat HTTP → one `tool_calls` bash, `finish_reason: tool_calls` |
| 7 | `openai/responses/tool-bash-with-reasoning-prefix` | Preset `think-tool-message`; 1 responses stream → reasoning preamble + `function_call` bash |
| 8 | `openai/responses/tool-bash-only` | Preset `tool-bash`; 1 responses stream → `function_call` bash without reasoning item |
| 9 | `anthropic/messages/think-tool-message-two-requests` | Preset `think-tool-message`; 2 `/v1/messages` HTTP → #1 thinking+tool_use, #2 text |
| 10 | `anthropic/messages/tool-bash-one-request` | Preset `tool-bash`; 1 `/v1/messages` HTTP → single `tool_use` block |

## Coverage

- **Breakpoint dequeue**: think collapse, one breakpoint per HTTP, never two tool_calls per response
- **Endpoints**: POST `/v1/chat/completions`, POST `/v1/responses` (stream), POST `/v1/messages`
- **OpenAI chat**: think merged into message content; think omitted from tool_call wire
- **OpenAI responses**: reasoning item before function_call (option B); codex tool remap
- **Anthropic messages**: native thinking blocks + single tool_use or text per response
- **Agent-events**: per-serve consumption order (think+tool vs message split across requests)
- **Regression guards**: `tool-bash` single-breakpoint on chat and responses encoders

## How to Run

```sh
# Run all breakpoint-queue tests (expect RED before implementation)
doctest vet ./agent/llm/llm-mock/tests/breakpoint-queue
doctest test ./agent/llm/llm-mock/tests/breakpoint-queue

# Run by group
doctest test ./agent/llm/llm-mock/tests/breakpoint-queue/dequeue/...
doctest test ./agent/llm/llm-mock/tests/breakpoint-queue/openai/...
doctest test ./agent/llm/llm-mock/tests/breakpoint-queue/anthropic/...

# Run a single leaf
doctest test ./agent/llm/llm-mock/tests/breakpoint-queue/dequeue/think-tool-message-two-requests
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
    BlockListener net.Listener
    ExpectedFallbackPort int
    Requests     []string
    Endpoint     string
    Method       string
    UseEnvConfig bool
    BinaryCmd []string
    BinaryEnv map[string]string
    EventsFile string
    AgentEventsFile string
    LogHTTPFile string
    ExpectedOutputContains string
    ExpectedExitCode int
    EventsInputJSONL string
    ServerDir string
    HTTPTimeout time.Duration
    MockEventsPreset string
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

    var configPath string
    if req.ConfigJSON != "" {
        configPath = filepath.Join(t.TempDir(), "llm-mock-config.json")
        if err := os.WriteFile(configPath, []byte(req.ConfigJSON), 0644); err != nil {
            return nil, fmt.Errorf("write config: %w", err)
        }
    }

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