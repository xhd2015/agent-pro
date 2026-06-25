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
- `Request.BinaryCmd` — command+args to spawn for binary integration tests.
- `Request.BinaryEnv` — extra environment variables for the spawned binary.
- `Request.EventsFile` — path to a JSON-lines events file written by the mock server.
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

// runBinary starts the mock server, spawns the external binary configured in
// req.BinaryCmd, waits for it to finish, queries the admin endpoint, then
// shuts down the server. It populates resp.Stdout (combined output), resp.ExitCode,
// and resp.Responses (admin query result).
func runBinary(t *testing.T, req *Request) (*Response, error) {
    resp := &Response{}

    // Set up events file if not specified
    if req.EventsFile == "" {
        req.EventsFile = filepath.Join(t.TempDir(), "llm-mock-events.jsonl")
    }

    // Write config JSON to temp file
    configPath := filepath.Join(t.TempDir(), "llm-mock-config.json")
    if err := os.WriteFile(configPath, []byte(req.ConfigJSON), 0644); err != nil {
        return nil, fmt.Errorf("write config: %w", err)
    }

    // Start the mock server
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

    // Read port from server stdout
    port, err := readPort(stdoutPipe)
    if err != nil {
        cancel()
        cmd.Wait()
        resp.Stderr = serverStderr.String()
        resp.Err = fmt.Errorf("read port: %w\nserver stderr: %s", err, resp.Stderr)
        return resp, nil
    }
    resp.Port = port

    // Spawn the external binary
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

    // Replace __MOCK_PORT__ placeholder with the actual port in env values
    portStr := fmt.Sprintf("%d", port)
    expandedEnv := make(map[string]string, len(req.BinaryEnv))
    for k, v := range req.BinaryEnv {
        expandedEnv[k] = strings.ReplaceAll(v, "__MOCK_PORT__", portStr)
    }

    // If PI_CODING_AGENT_DIR is set, also replace __MOCK_PORT__ in models.json
    // on disk (the file was written in Setup before the port was known).
    if piDir, ok := expandedEnv["PI_CODING_AGENT_DIR"]; ok {
        modelsPath := filepath.Join(piDir, "models.json")
        data, readErr := os.ReadFile(modelsPath)
        if readErr == nil {
            updated := strings.ReplaceAll(string(data), "__MOCK_PORT__", portStr)
            os.WriteFile(modelsPath, []byte(updated), 0644)
        }
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

    // Shut down the mock server before reading events (ensures file is flushed)
    cancel()
    cmd.Wait()

    // Read events from the events file
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
