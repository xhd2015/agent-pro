# Scenario

**Feature**: breakpoint dequeue and API encoders on llm-mock HTTP server

```
# seed genQueue from preset, dequeue to breakpoint per HTTP serve
config loader -> genQueue [think*, breakpoint] -> DequeueToBreakpoint -> encoder -> JSON/SSE response

# agent-events mirror consumed slice
HTTP client <- llm-mock (--agent-events-file JSONL per serve)
```

## Preconditions

- Repository contains `agent/llm/llm-mock` (built as server binary).
- Tests build the binary, start server with `--mock-events-preset`, send sequential HTTP POSTs.
- Empty `exchanges: []` config unless a leaf overrides.
- Implementer adds `two-tool-message` preset (`tool_call` bash, `tool_call` read, `message`) for leaf #2.

## Steps

1. Build `agent/llm/llm-mock` binary.
2. Write temp JSON config when `Request.ConfigJSON` is set.
3. Start server with `--agent-events-file` and optional `--mock-events-preset`.
4. Parse listening port from server stdout.
5. Send ordered HTTP requests to `Request.Endpoint`.
6. Read agent-events JSONL and collect responses.

## Context

- `Request.Endpoint` — `/v1/chat/completions`, `/v1/responses`, or `/v1/messages`.
- `Request.MockEventsPreset` — seeds `genQueue` after prefix exhaustion.
- `Request.Requests` — ordered JSON bodies for sequential HTTP POSTs.
- `Response.AgentEventsLines` — served AgentEvent JSONL (consumption proof per HTTP serve).
- Preset text constants live in `mockpreset` package (`preset:think:…`, `preset:message:…`).

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

func Setup(t *testing.T, req *Request) error {
    _ = assertSuccess
    _ = assertContains
    _ = assertNotContains
    _ = parseJSON
    _ = parseSSEEvents
    _ = readAgentEventLines
    _ = parseAgentEventMaps
    _ = anthropicContentBlocks
    _ = countAgentEventsByType

    req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
    if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
        return fmt.Errorf("repo root not found: %w", err)
    }
    req.BinaryPath = filepath.Join(t.TempDir(), "llm-mock")
    if req.Port == 0 {
        req.Port = 8080
    }
    if req.Endpoint == "" {
        req.Endpoint = "/v1/chat/completions"
    }
    if req.Method == "" {
        req.Method = "POST"
    }

    build := exec.Command("go", "build", "-o", req.BinaryPath, "./agent/llm/llm-mock")
    build.Dir = req.RepoRoot
    if out, err := build.CombinedOutput(); err != nil {
        return fmt.Errorf("build llm-mock: %w\n%s", err, string(out))
    }
    return nil
}

func readPort(stdout io.Reader) (int, error) {
    scanner := bufio.NewScanner(stdout)
    type result struct {
        port int
        err  error
    }
    ch := make(chan result, 1)
    var mu sync.Mutex
    var lines []string

    go func() {
        for scanner.Scan() {
            line := scanner.Text()
            mu.Lock()
            lines = append(lines, line)
            mu.Unlock()
            trimmed := strings.TrimSpace(line)
            if strings.HasPrefix(trimmed, ":") {
                p := 0
                if _, err := fmt.Sscanf(trimmed, ":%d", &p); err == nil && p > 0 {
                    ch <- result{port: p}
                    return
                }
            }
            p := 0
            if _, err := fmt.Sscanf(trimmed, "%d", &p); err == nil && p > 0 && p < 65536 && len(strings.Fields(trimmed)) == 1 {
                ch <- result{port: p}
                return
            }
        }
        mu.Lock()
        allLines := strings.Join(lines, "\n")
        mu.Unlock()
        ch <- result{err: fmt.Errorf("no port found in stdout:\n%s", allLines)}
    }()

    select {
    case r := <-ch:
        return r.port, r.err
    case <-time.After(10 * time.Second):
        mu.Lock()
        allLines := strings.Join(lines, "\n")
        mu.Unlock()
        return 0, fmt.Errorf("timeout waiting for server port\nstdout so far:\n%s", allLines)
    }
}

func makeHTTPRequestWithTimeout(baseURL, endpoint, method, bodyJSON string, timeout time.Duration) (HTTPResponse, error) {
    url := baseURL + endpoint
    var bodyReader io.Reader
    if bodyJSON != "" {
        bodyReader = strings.NewReader(bodyJSON)
    }

    httpReq, err := http.NewRequest(method, url, bodyReader)
    if err != nil {
        return HTTPResponse{}, fmt.Errorf("create request: %w", err)
    }
    if bodyJSON != "" {
        httpReq.Header.Set("Content-Type", "application/json")
    }

    client := &http.Client{Timeout: timeout}
    httpResp, err := client.Do(httpReq)
    if err != nil {
        return HTTPResponse{StatusCode: 0, Body: "", Headers: nil}, nil
    }
    defer httpResp.Body.Close()

    bodyBytes, err := io.ReadAll(httpResp.Body)
    if err != nil {
        return HTTPResponse{StatusCode: httpResp.StatusCode}, fmt.Errorf("read body: %w", err)
    }

    headers := make(map[string]string)
    for k, v := range httpResp.Header {
        headers[k] = strings.Join(v, ", ")
    }

    return HTTPResponse{
        StatusCode: httpResp.StatusCode,
        Body:       string(bodyBytes),
        Headers:    headers,
    }, nil
}

func runBinary(t *testing.T, req *Request) (*Response, error) {
    resp := &Response{}

    if req.EventsFile == "" {
        req.EventsFile = filepath.Join(t.TempDir(), "llm-mock-events.jsonl")
    }

    configPath := filepath.Join(t.TempDir(), "llm-mock-config.json")
    if err := os.WriteFile(configPath, []byte(req.ConfigJSON), 0644); err != nil {
        return nil, fmt.Errorf("write config: %w", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    args := []string{}
    if !req.UseEnvConfig {
        args = append(args, "--config", configPath)
    }
    args = append(args, "--events-file", req.EventsFile)

    cmd := exec.CommandContext(ctx, req.BinaryPath, args...)
    cmd.Dir = t.TempDir()

    env := os.Environ()
    if req.UseEnvConfig {
        env = append(env, "LLM_MOCK_CONFIG="+configPath)
    }
    cmd.Env = env

    var serverStderr bytes.Buffer
    cmd.Stderr = &serverStderr

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
        cmd.Wait()
        resp.Stderr = serverStderr.String()
        resp.Err = fmt.Errorf("read port: %w\nserver stderr: %s", err, resp.Stderr)
        return resp, nil
    }
    resp.Port = port

    if len(req.BinaryCmd) == 0 {
        cancel()
        cmd.Wait()
        return nil, fmt.Errorf("BinaryCmd is empty in runBinary (should not happen)")
    }

    binaryName := req.BinaryCmd[0]
    binaryArgs := req.BinaryCmd[1:]

    binaryCtx, binaryCancel := context.WithTimeout(ctx, 50*time.Second)
    defer binaryCancel()

    binaryCmd := exec.CommandContext(binaryCtx, binaryName, binaryArgs...)
    binaryCmd.Dir = t.TempDir()

    portStr := fmt.Sprintf("%d", port)
    expandedEnv := make(map[string]string, len(req.BinaryEnv))
    for k, v := range req.BinaryEnv {
        expandedEnv[k] = strings.ReplaceAll(v, "__MOCK_PORT__", portStr)
    }
    binaryCmd.Env = envWithOverrides(os.Environ(), expandedEnv)

    var combinedOut bytes.Buffer
    binaryCmd.Stdout = &combinedOut
    binaryCmd.Stderr = &combinedOut

    runErr := binaryCmd.Run()
    resp.Stdout = combinedOut.String()

    if runErr != nil {
        var exitErr *exec.ExitError
        if errors.As(runErr, &exitErr) {
            resp.ExitCode = exitErr.ExitCode()
        } else {
            resp.ExitCode = -1
            resp.Err = runErr
        }
    }

    cancel()
    cmd.Wait()

    eventsData, readErr := os.ReadFile(req.EventsFile)
    if readErr != nil {
        resp.Responses = append(resp.Responses, HTTPResponse{
            StatusCode: 0,
            Body:       fmt.Sprintf("events file read error: %v", readErr),
        })
    } else {
        resp.Responses = append(resp.Responses, HTTPResponse{
            StatusCode: 200,
            Body:       string(eventsData),
        })
    }

    return resp, nil
}

func assertSuccess(t *testing.T, resp *Response) {
    t.Helper()
    if resp.Err != nil && resp.ExitCode == 0 {
        t.Fatalf("run failed: %v\nstderr: %s", resp.Err, resp.Stderr)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
    }
}

func assertContains(t *testing.T, got string, want string) {
    t.Helper()
    if !strings.Contains(got, want) {
        t.Fatalf("missing %q in:\n%s", want, got)
    }
}

func assertNotContains(t *testing.T, got string, want string) {
    t.Helper()
    if strings.Contains(got, want) {
        t.Fatalf("unexpected %q in:\n%s", want, got)
    }
}

func envWithOverrides(base []string, overrides map[string]string) []string {
    merged := make(map[string]string, len(base)+len(overrides))
    for _, kv := range base {
        key, value, ok := strings.Cut(kv, "=")
        if !ok || key == "" {
            continue
        }
        merged[key] = value
    }
    for key, value := range overrides {
        merged[key] = value
    }
    out := make([]string, 0, len(merged))
    for key, value := range merged {
        out = append(out, key+"="+value)
    }
    return out
}

func parseJSON(t *testing.T, text string) map[string]any {
    t.Helper()
    var obj map[string]any
    if err := json.Unmarshal([]byte(text), &obj); err != nil {
        t.Fatalf("invalid JSON: %v\n%s", err, text)
    }
    return obj
}

func parseSSEEvents(t *testing.T, body string) []map[string]any {
    t.Helper()
    var events []map[string]any
    lines := strings.Split(body, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if !strings.HasPrefix(line, "data: ") {
            continue
        }
        data := strings.TrimPrefix(line, "data: ")
        if data == "[DONE]" {
            continue
        }
        var obj map[string]any
        if err := json.Unmarshal([]byte(data), &obj); err != nil {
            t.Fatalf("invalid SSE JSON: %v\n%s", err, data)
        }
        events = append(events, obj)
    }
    return events
}

func readAgentEventLines(text string) []string {
    var lines []string
    for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
        line = strings.TrimSpace(line)
        if line != "" {
            lines = append(lines, line)
        }
    }
    return lines
}

func parseAgentEventMaps(lines []string) ([]map[string]any, error) {
    out := make([]map[string]any, 0, len(lines))
    for i, line := range lines {
        var ev map[string]any
        if err := json.Unmarshal([]byte(line), &ev); err != nil {
            return nil, fmt.Errorf("line %d: invalid JSONL: %w\n%s", i+1, err, line)
        }
        typ, _ := ev["type"].(string)
        if typ == "" {
            return nil, fmt.Errorf("line %d: missing type in %#v", i+1, ev)
        }
        out = append(out, ev)
    }
    return out, nil
}

func countAgentEventsByType(events []map[string]any) map[string]int {
    counts := make(map[string]int)
    for _, ev := range events {
        typ, _ := ev["type"].(string)
        counts[typ]++
    }
    return counts
}

func anthropicContentBlocks(t *testing.T, body string) []map[string]any {
    t.Helper()
    obj := parseJSON(t, body)
    content, ok := obj["content"].([]any)
    if !ok {
        t.Fatalf("anthropic response missing content[]: %s", body)
    }
    blocks := make([]map[string]any, 0, len(content))
    for i, item := range content {
        block, ok := item.(map[string]any)
        if !ok {
            t.Fatalf("content[%d] not object", i)
        }
        blocks = append(blocks, block)
    }
    return blocks
}
```