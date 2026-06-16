## Preconditions
- The repository contains the `agent/llm/llm-mock` package (to be built as the server binary).
- Each test builds the server binary, writes a temp JSON config, starts the server,
  sends HTTP requests, and returns parsed responses.
- The server prints the listening address to stdout on startup (e.g., `:8080`).
- The server shuts down on context cancellation.

## Steps
1. Build `agent/llm/llm-mock` binary.
2. Write the test-specific JSON config to a temp file.
3. Start the server with `--config <path>` (or `LLM_MOCK_CONFIG` env var).
4. Parse the listening port from server stdout.
5. Send HTTP requests to the server.
6. Collect responses, stderr, and exit status.
7. Cancel the server context and wait for shutdown.

## Context
- `Request.ConfigJSON` — the JSON configuration content for the mock server.
- `Request.BlockPort` — if >0, bind this port before starting the server (for fallback testing).
- `Request.Requests` — ordered slice of JSON request bodies to POST to the chat completions endpoint.
- `Request.Endpoint` — URL path (e.g., `/v1/chat/completions`, `/v1/models`).
- `Request.Method` — HTTP method, defaults to POST.
- `Request.UseEnvConfig` — if true, pass config via `LLM_MOCK_CONFIG` env var instead of `--config`.
- `Response.Responses` — ordered slice of HTTP responses, one per request.
- `Response.Port` — the actual port the server is listening on.

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
    Requests     []string
    Endpoint     string
    Method       string
    UseEnvConfig bool
}

type Response struct {
    Responses []HTTPResponse
    ExitCode  int
    Stdout    string
    Stderr    string
    Port      int
    Err       error
}

func Setup(t *testing.T, req *Request) error {
    _ = assertSuccess
    _ = assertContains
    _ = assertNotContains
    _ = parseJSON
    _ = parseSSEEvents

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

func Run(t *testing.T, req *Request) (*Response, error) {
    resp := &Response{}

    // Write config JSON to temp file
    configPath := filepath.Join(t.TempDir(), "llm-mock-config.json")
    if err := os.WriteFile(configPath, []byte(req.ConfigJSON), 0644); err != nil {
        return nil, fmt.Errorf("write config: %w", err)
    }

    // If BlockPort is set, bind it to force port fallback
    var blockListener net.Listener
    if req.BlockPort > 0 {
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

    args := []string{}
    if !req.UseEnvConfig {
        args = append(args, "--config", configPath)
    }

    cmd := exec.CommandContext(ctx, req.BinaryPath, args...)
    cmd.Dir = t.TempDir()

    env := os.Environ()
    if req.UseEnvConfig {
        env = append(env, "LLM_MOCK_CONFIG="+configPath)
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
        cmd.Wait()
        resp.Stderr = stderrBuf.String()
        resp.Err = fmt.Errorf("read port: %w\nstderr: %s", err, resp.Stderr)
        return resp, nil
    }
    resp.Port = port

    // Make HTTP requests
    baseURL := fmt.Sprintf("http://localhost:%d", port)
    for i, bodyJSON := range req.Requests {
        httpResp, err := makeHTTPRequest(baseURL, req.Endpoint, req.Method, bodyJSON)
        if err != nil {
            cancel()
            cmd.Wait()
            resp.Stderr = stderrBuf.String()
            resp.Err = fmt.Errorf("HTTP request %d: %w", i, err)
            return resp, nil
        }
        resp.Responses = append(resp.Responses, httpResp)
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

// readPort reads from the server's stdout pipe looking for a port announcement.
// Expects a line containing a port in the format ":8080" or "8080".
func readPort(stdout io.Reader) (int, error) {
    scanner := bufio.NewScanner(stdout)
    // Use a buffered approach: read lines for up to 5 seconds to find the port
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
            // Look for port pattern: :NNNNN or just NNNNN
            trimmed := strings.TrimSpace(line)
            // Try ":8080" format
            if strings.HasPrefix(trimmed, ":") {
                p := 0
                if _, err := fmt.Sscanf(trimmed, ":%d", &p); err == nil && p > 0 {
                    ch <- result{port: p}
                    return
                }
            }
            // Try bare port number on its own line
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

// makeHTTPRequest sends an HTTP request and returns the parsed response.
func makeHTTPRequest(baseURL, endpoint, method, bodyJSON string) (HTTPResponse, error) {
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

    client := &http.Client{Timeout: 10 * time.Second}
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

func parseJSON(t *testing.T, text string) map[string]any {
    t.Helper()
    var obj map[string]any
    if err := json.Unmarshal([]byte(text), &obj); err != nil {
        t.Fatalf("invalid JSON: %v\n%s", err, text)
    }
    return obj
}

// parseSSEEvents extracts "data:" lines from an SSE stream body.
// Returns the parsed JSON objects (does not parse [DONE] as JSON).
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
```
