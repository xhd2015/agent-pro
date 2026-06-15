## Preconditions
- Each leaf provides structured input data in request fields.
- `Run` dispatches based on which fields are set: `Events` → ToCodex/ToOpencode, `Value` → marshal, `Output` → pass-through.

## Steps
1. If `req.Events` is set, call `codex_types.ToCodex` or `opencode_types.ToOpencode` and marshal.
2. If `req.Value` is set, marshal it to JSON.
3. Otherwise, pass through `req.Output`.

```go
import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	codex_types "github.com/xhd2015/agent-pro/agent/event/codex_types"
	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	Events     []types.AgentEvent
	Target     string // "opencode"→ToOpencode, "crush"→ToCrush, "from_crush"→FromCrush, "crush_server"→StartCrushServer; default→ToCodex
	SessionID  string
	Value      any
	Output     string
	CrushInput string // raw JSON for FromCrush parsing
	HostPort   int    // crush server HTTP port (0 = auto-assign)
	Prompt     string // prompt text for crush server
	ModelName  string // model override (default: "deepseek-v4-pro")
	CrushPath  string // path to crush binary (default: LookPath("crush"))
}

type Response struct {
	Output string
}

func Setup(t *testing.T, req *Request) error {
	_ = assertContains
	_ = assertNotContains
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var output string
	if req.Target == "from_crush" && req.CrushInput != "" {
		var crushEvents []crush_types.Event
		if err := json.Unmarshal([]byte(req.CrushInput), &crushEvents); err != nil {
			return &Response{Output: ""}, nil
		}
		result := crush_types.FromCrush(crushEvents, req.SessionID)
		data, _ := json.Marshal(result)
		output = string(data)
	} else if len(req.Events) > 0 {
		if req.Target == "opencode" {
			result := opencode_types.ToOpencode(req.Events, req.SessionID)
			data, _ := json.Marshal(result)
			output = string(data)
		} else if req.Target == "crush" {
			result := crush_types.ToCrush(req.Events, req.SessionID)
			data, _ := json.Marshal(result)
			output = string(data)
		} else {
			result := codex_types.ToCodex(req.Events)
			data, _ := json.Marshal(result)
			output = string(data)
		}
	} else if req.Value != nil {
		data, _ := json.Marshal(req.Value)
		output = string(data)
	} else if req.Target == "crush_server" {
		return runCrushServer(t, req)
	} else {
		output = req.Output
	}
	return &Response{Output: output}, nil
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

func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("rand.Read: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

func runCrushServer(t *testing.T, req *Request) (*Response, error) {
	if os.Getenv("CRUSH_SKIP_INTEGRATION") == "1" {
		t.Skip("CRUSH_SKIP_INTEGRATION=1")
		return nil, nil
	}

	crushPath := req.CrushPath
	if crushPath == "" {
		var err error
		crushPath, err = exec.LookPath("crush")
		if err != nil {
			t.Skip("crush binary not found on PATH")
			return nil, nil
		}
	} else {
		if _, err := os.Stat(crushPath); err != nil {
			return nil, fmt.Errorf("crush binary not found at %s: %w", crushPath, err)
		}
	}

	port := req.HostPort
	if port == 0 {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, fmt.Errorf("failed to pick port: %w", err)
		}
		port = listener.Addr().(*net.TCPAddr).Port
		listener.Close()
	}

	addr := fmt.Sprintf("tcp://localhost:%d", port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, crushPath, "server", "--host", addr)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start crush server: %w", err)
	}
	defer func() {
		cancel()
		cmd.Wait()
	}()

	healthURL := fmt.Sprintf("http://localhost:%d/v1/health", port)
	healthCtx, healthCancel := context.WithTimeout(ctx, 30*time.Second)
	defer healthCancel()
	healthOK := false
	for !healthOK {
		select {
		case <-healthCtx.Done():
			return nil, fmt.Errorf("crush server health check timed out")
		default:
		}
		resp, err := http.Get(healthURL)
		if err == nil && resp != nil && resp.StatusCode == 200 {
			resp.Body.Close()
			healthOK = true
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	clientID := newUUID()

	tmpDir, err := os.MkdirTemp("", "crush-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	wsBody, _ := json.Marshal(map[string]string{"path": tmpDir, "client_id": clientID})
	wsResp, err := http.Post(baseURL+"/v1/workspaces", "application/json", bytes.NewReader(wsBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	defer wsResp.Body.Close()
	var ws struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(wsResp.Body).Decode(&ws); err != nil {
		return nil, fmt.Errorf("failed to decode workspace response: %w", err)
	}
	if ws.ID == "" {
		return nil, fmt.Errorf("workspace ID is empty")
	}

	sessResp, err := http.Post(baseURL+"/v1/workspaces/"+ws.ID+"/sessions", "application/json", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer sessResp.Body.Close()
	var sess struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(sessResp.Body).Decode(&sess); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %w", err)
	}
	if sess.ID == "" {
		return nil, fmt.Errorf("session ID is empty")
	}

	sseURL := fmt.Sprintf("%s/v1/workspaces/%s/events?client_id=%s", baseURL, ws.ID, clientID)
	sseReq, err := http.NewRequestWithContext(ctx, "GET", sseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE request: %w", err)
	}
	sseReq.Header.Set("Accept", "text/event-stream")
	sseReq.Header.Set("Cache-Control", "no-cache")
	sseReq.Header.Set("Connection", "keep-alive")
	sseResp, err := http.DefaultClient.Do(sseReq)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}
	defer sseResp.Body.Close()

	sendBody := fmt.Sprintf(`{"session_id":"%s","prompt":"%s"}`,
		escapeJSON(sess.ID), escapeJSON(req.Prompt))
	sendResp, err := http.Post(baseURL+"/v1/workspaces/"+ws.ID+"/agent", "application/json", strings.NewReader(sendBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	sendResp.Body.Close()

	runCtx, runCancel := context.WithTimeout(ctx, 120*time.Second)
	defer runCancel()

	var events []crush_types.Event
	scanner := bufio.NewScanner(sseResp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

		var outer struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal([]byte(dataStr), &outer); err != nil {
			continue
		}

		var inner struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(outer.Payload, &inner); err != nil {
			continue
		}

		event := crush_types.Event{
			Type:    crush_types.EventType(outer.Type),
			Payload: inner.Payload,
		}
		events = append(events, event)

		if outer.Type == string(crush_types.EventRunComplete) {
			break
		}

		select {
		case <-runCtx.Done():
			return nil, fmt.Errorf("timed out waiting for run_complete")
		default:
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SSE events: %w", err)
	}

	result := crush_types.FromCrush(events, req.SessionID)
	data, _ := json.Marshal(result)
	return &Response{Output: string(data)}, nil
}
```
