# Scenario

**Feature**: `agent-run tty` subcommands — status, attach, send for live TTY sessions

```
# status reads registry entry + probes TCP / ptywrap WebSocket
agent-run tty status <session-id> -> registry file -> human-readable/JSON output

# attach connects via ptyclient.Attach
agent-run tty attach <session-id> -> registry lookup -> ptywrap WS connect

# send injects prompt into live terminal and captures response
agent-run tty send <session-id> "msg" -> registry lookup -> WS inject prompt -> capture response -> append events
```

## Preconditions

- Repository contains `cmd/agent-run` (build may fail until tty subcommands are implemented).
- Each test uses isolated `AGENT_RUN_HOME=filepath.Join(t.TempDir(), ".agent-run")`.
- Tests use mock registry JSON files and optionally a fake in-process ptywrap HTTP+WebSocket server.
- No real Codex or Grok CLI is required.

## Steps

1. Root `Setup` builds `agent-run`, sets `AGENT_RUN_HOME`.
2. Grouping `Setup` writes a mock registry entry and/or starts a fake ptywrap server.
3. Leaf `Setup` finalizes `req.Args` for the specific CLI invocation.
4. `Run` executes `agent-run tty <cmd>` (or `agent-run attach`) and collects output.
5. Leaf `Assert` checks exit code, stdout/stderr content, JSON structure, or attach probe result.

```go
import (
	"runtime"
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

	"github.com/gorilla/websocket"
	"github.com/xhd2015/doctest/session"
)

var fakePTYWrapUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.AgentRun, "./agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	return nil
}

func registryDirFor(home, dirName string) string {
	return filepath.Join(home, dirName)
}

func registryPathFor(home, dirName, sessionID string) string {
	return filepath.Join(registryDirFor(home, dirName), sessionID+".json")
}

func writeMockRegistryEntry(t *testing.T, req *Request) {
	t.Helper()
	dirName := req.RegistryDir
	if dirName == "" {
		dirName = "grok-tty-registry"
	}
	entry := req.RegistryEntryJSON
	if entry == "" {
		entry = defaultRegistryEntryJSON(req.RegistrySessionID, req.FakePTYWrapPort)
	}
	sessionID := req.RegistrySessionID
	if sessionID == "" {
		sessionID = "session-1"
		req.RegistrySessionID = sessionID
	}
	dir := registryDirFor(req.Home, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	path := registryPathFor(req.Home, dirName, sessionID)
	if err := os.WriteFile(path, []byte(entry), 0644); err != nil {
		t.Fatalf("write registry entry: %v", err)
	}
}

func defaultRegistryEntryJSON(sessionID string, port int) string {
	if sessionID == "" {
		sessionID = "session-1"
	}
	listenAddr := fmt.Sprintf("127.0.0.1:%d", port)
	if port <= 0 {
		listenAddr = "127.0.0.1:12345"
	}
	entry := RegistryEntryData{
		SessionID:  sessionID,
		ListenAddr: listenAddr,
		PID:        12345,
		CreatedAt:  "2026-07-03T12:00:00Z",
	}
	b, _ := json.Marshal(entry)
	return string(b)
}

func startFakePTYWrapServer(t *testing.T, req *Request) {
	t.Helper()
	if !req.StartFakePTYWrap {
		return
	}
	mux := http.NewServeMux()

	var inputReceivedMu sync.Mutex
	var inputReceived []string

	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[{"id":"session-1"}]`)
	})

	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := fakePTYWrapUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		scrollback := req.FakePTYWrapScrollback
		if scrollback == "" {
			// Modern boxed composer (section judge). Legacy "Grok ›"/">" is not writable-idle.
			scrollback = "" +
				"GROK_TTY_BANNER\n" +
				"prompt text\n" +
				"Response: hello world\n" +
				" ⎇ master worktree ~/.wrk/… 1K / 10K\n" +
				"    Worked for 1.0s                                        stop  [hooks: 1]\n" +
				" ╭--------------------------------------------------------------------------╮\n" +
				" │ ❯                                                                        │\n" +
				" ╰----------------------------------------- Grok 4.5 (high) · always-approve -╯\n" +
				" Shift+Tab:mode  │  Ctrl+.:shortcuts\n"
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(scrollback)); err != nil {
			return
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			inputReceivedMu.Lock()
			inputReceived = append(inputReceived, string(msg))
			inputReceivedMu.Unlock()
			if req.FakePTYInputReceived != nil {
				req.FakePTYInputReceived <- string(msg)
			}
		}
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake ptywrap listen: %v", err)
	}
	req.FakePTYWrapPort = ln.Addr().(*net.TCPAddr).Port

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})

	if req.FakePTYReady != nil {
		close(req.FakePTYReady)
	}
}

func waitForPortOpen(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s to be reachable", addr)
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	if len(args) == 0 {
		args = req.Args
	}
	return execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
}

func runStatusJSON(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp, err := runAgentRun(t, req, req.Args...)
	if err != nil {
		return resp, err
	}
	if resp.ExitCode != 0 {
		return resp, nil
	}
	var obj map[string]any
	if jsonErr := json.Unmarshal([]byte(resp.Stdout), &obj); jsonErr != nil {
		return resp, fmt.Errorf("parse JSON stdout: %w\nstdout:\n%s", jsonErr, resp.Stdout)
	}
	resp.JSONBody = obj
	return resp, nil
}

func runAttachProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	timeout := req.AttachTimeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, req.Args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		resp.AttachProbeOK = true
		return resp, nil
	}
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			resp.AttachProbeOK = true
			return resp, nil
		}
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runSendProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, req.Args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("expected exit code %d, got %d, stderr:\n%s", want, resp.ExitCode, resp.Stderr)
	}
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	assertExitCode(t, resp, 0)
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertOutput(t *testing.T, resp *Response, stream string, contains ...string) {
	t.Helper()
	var got string
	switch stream {
	case "stdout":
		got = resp.Stdout
	case "stderr":
		got = resp.Stderr
	default:
		t.Fatalf("unknown stream %q (want stdout or stderr)", stream)
	}
	for _, want := range contains {
		assertContains(t, got, want)
	}
}

func findAgentProRoot(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) == "module github.com/xhd2015/agent-pro" {
					return dir, nil
				}
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find agent-pro module root above %s", start)
		}
	}
}

```
