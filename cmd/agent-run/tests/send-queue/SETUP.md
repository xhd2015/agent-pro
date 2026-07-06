# Scenario

**Feature**: agent-run send queue with session-local message ids

```
agent-run send <session-id> "msg" -> agentsend.Enqueue -> stdout msg_N -> drainer -> ttywatch.SendMessage
agent-run msg status|cancel <session-id>/msg_N -> agentsend.MessageStatus / Cancel
```

## Preconditions

- Package `pkgs/agentsend` implements per-session JSONL queue (may not exist yet — tests RED).
- Repository contains `cmd/agent-run` with `send` / `tty send` wired to agentsend.
- Each test uses isolated `AGENT_RUN_HOME=filepath.Join(t.TempDir(), ".agent-run")`.
- Live-session tests set `AGENT_RUN_ENABLE_STUB_TTY=1` and start background stub-tty.
- Queue files live at `AGENT_RUN_HOME/send-queue/<runner>/<terminal_session_id>.jsonl`.

## Steps

1. Root `Setup` builds `agent-run`, sets `AGENT_RUN_HOME`, enables stub-tty when needed.
2. Grouping `Setup` starts background stub-tty with scenario JSON (idle / busy / busy-then-idle).
3. Leaf `Setup` sets `req.Operation`, `req.Action`, send flags, and messages.
4. `Run` executes CLI commands and captures stdout/stderr, timing, injection order, queue state.
5. Leaf `Assert` checks exit codes, id stdout contract, stderr errors, FIFO order, queue file.

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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var stubSessionIDRe = regexp.MustCompile(`stub-tty:\s*(session-\d+)`)
var msgIDLineRe = regexp.MustCompile(`^msg_\d+$`)

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command("go", "build", "-o", req.AgentRun, "./cmd/agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	req.Env = append(req.Env, "AGENT_RUN_ENABLE_STUB_TTY=1")
	if req.RunnerID == "" {
		req.RunnerID = "stub-tty"
	}
	return nil
}

func runSendQueueOp(t *testing.T, req *Request) (*Response, error) {
	switch req.Operation {
	case "enqueue":
		return runEnqueueOp(t, req)
	case "wait":
		return runWaitOp(t, req)
	case "fifo":
		return runFifoOp(t, req)
	case "cancel":
		return runCancelOp(t, req)
	case "errors":
		return runErrorsOp(t, req)
	case "alias":
		return runAliasOp(t, req)
	default:
		return nil, fmt.Errorf("unknown operation %q", req.Operation)
	}
}

func runEnqueueOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Action {
	case "first-send-prints-msg-1":
		startIdleStubSession(t, req)
		capture := startInjectionCapture(t, req)
		defer injectionCaptureStop(capture)
		sendResp, err := execSend(t, req, req.SendMessage)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		resp.MsgID = strings.TrimSpace(resp.Stdout)
		resp.InjectedMessages = injectionCaptureWaitFor(capture, req.SendMessage, 10*time.Second)
		return resp, nil
	case "second-send-prints-msg-2":
		startIdleStubSession(t, req)
		first, err := execSend(t, req, "first-message", "--no-wait")
		if err != nil {
			return resp, err
		}
		if first.ExitCode != 0 {
			return resp, fmt.Errorf("first send failed: exit=%d stderr=%s", first.ExitCode, first.Stderr)
		}
		second, err := execSend(t, req, req.SendMessage)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, second)
		resp.SecondStdout = first.Stdout
		resp.SecondMsgID = strings.TrimSpace(second.Stdout)
		return resp, nil
	case "no-wait-prints-id-immediately":
		startBusyStubSession(t, req)
		start := time.Now()
		sendResp, err := execSend(t, req, req.SendMessage, "--no-wait")
		resp.SendDuration = time.Since(start)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		resp.MsgID = strings.TrimSpace(resp.Stdout)
		return resp, nil
	case "max-wait-prints-id-before-wait":
		startBusyStubSession(t, req)
		idLine, latency, sendResp, err := execSendCaptureFirstStdoutLine(t, req, 12*time.Second, req.SendMessage, "--max-wait", "10s")
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		resp.MsgID = idLine
		resp.IdLineLatency = latency
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown enqueue action %q", req.Action)
	}
}

func runWaitOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Action {
	case "default-waits-until-writable-then-delivers":
		req.StubScenarioJSON = stubScenarioBusyThenIdleJSON()
		startIdleStubSession(t, req)
		capture := startInjectionCapture(t, req)
		defer injectionCaptureStop(capture)
		start := time.Now()
		sendResp, err := execSend(t, req, req.SendMessage)
		resp.SendDuration = time.Since(start)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		resp.MsgID = strings.TrimSpace(resp.Stdout)
		resp.InjectedMessages = injectionCaptureWaitFor(capture, req.SendMessage, 20*time.Second)
		return resp, nil
	case "max-wait-times-out-removes-message":
		startBusyStubSession(t, req)
		sendResp, err := execSend(t, req, req.SendMessage, "--max-wait", "2s")
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		resp.MsgID = strings.TrimSpace(resp.Stdout)
		resp.QueueFilePath = queueFilePath(req.Home, req.RunnerID, req.TerminalSessionID)
		resp.QueueHasMsgID = queueContainsMsgID(t, resp.QueueFilePath, resp.MsgID)
		return resp, nil
	case "no-wait-returns-before-delivery":
		startBusyStubSession(t, req)
		capture := startInjectionCapture(t, req)
		defer injectionCaptureStop(capture)
		start := time.Now()
		sendResp, err := execSend(t, req, req.SendMessage, "--no-wait")
		resp.SendDuration = time.Since(start)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		resp.MsgID = strings.TrimSpace(resp.Stdout)
		time.Sleep(400 * time.Millisecond)
		resp.InjectedMessages = injectionCaptureSnapshot(capture)
		resp.QueueFilePath = queueFilePath(req.Home, req.RunnerID, req.TerminalSessionID)
		resp.QueueHasMsgID = queueContainsMsgID(t, resp.QueueFilePath, resp.MsgID)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown wait action %q", req.Action)
	}
}

func runFifoOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	startIdleStubSession(t, req)
	capture := startInjectionCapture(t, req)
	defer injectionCaptureStop(capture)
	msgA := "fifo-message-A"
	msgB := "fifo-message-B"
	if _, err := execSend(t, req, msgA, "--no-wait"); err != nil {
		return resp, err
	}
	if _, err := execSend(t, req, msgB, "--no-wait"); err != nil {
		return resp, err
	}
	sendResp, err := execSend(t, req, "fifo-trigger")
	if err != nil {
		return resp, err
	}
	copySendResp(resp, sendResp)
	order := injectionCaptureWaitForOrdered(capture, []string{msgA, msgB}, 15*time.Second)
	resp.InjectedMessages = order
	return resp, nil
}

func runCancelOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Action {
	case "cancel-pending-message":
		startBusyStubSession(t, req)
		capture := startInjectionCapture(t, req)
		defer injectionCaptureStop(capture)
		sendResp, err := execSend(t, req, req.SendMessage, "--no-wait")
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		msgID := strings.TrimSpace(resp.Stdout)
		ref := sessionMessageRef(req.TerminalSessionID, msgID)
		statusBefore, err := execAgentRun(t, req, "msg", "status", ref)
		if err != nil {
			return resp, err
		}
		resp.StatusBeforeStdout = statusBefore.Stdout
		cancelResp, err := execAgentRun(t, req, "msg", "cancel", ref)
		if err != nil {
			return resp, err
		}
		resp.CancelStdout = cancelResp.Stdout
		resp.CancelStderr = cancelResp.Stderr
		resp.CancelExitCode = cancelResp.ExitCode
		statusAfter, err := execAgentRun(t, req, "msg", "status", ref)
		if err != nil {
			return resp, err
		}
		resp.StatusAfterStdout = statusAfter.Stdout
		time.Sleep(500 * time.Millisecond)
		resp.InjectedMessages = injectionCaptureSnapshot(capture)
		resp.QueueFilePath = queueFilePath(req.Home, req.RunnerID, req.TerminalSessionID)
		resp.QueueHasMsgID = queueContainsMsgID(t, resp.QueueFilePath, msgID)
		return resp, nil
	case "cancel-unknown-id":
		startIdleStubSession(t, req)
		cancelResp, err := execAgentRun(t, req, "msg", "cancel", sessionMessageRef(req.TerminalSessionID, "msg_9999"))
		if err != nil {
			return resp, err
		}
		resp.CancelStdout = cancelResp.Stdout
		resp.CancelStderr = cancelResp.Stderr
		resp.CancelExitCode = cancelResp.ExitCode
		return resp, nil
	case "cancel-after-delivered-fails":
		startIdleStubSession(t, req)
		sendResp, err := execSend(t, req, req.SendMessage)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		msgID := strings.TrimSpace(resp.Stdout)
		cancelResp, err := execAgentRun(t, req, "msg", "cancel", sessionMessageRef(req.TerminalSessionID, msgID))
		if err != nil {
			return resp, err
		}
		resp.CancelStdout = cancelResp.Stdout
		resp.CancelStderr = cancelResp.Stderr
		resp.CancelExitCode = cancelResp.ExitCode
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown cancel action %q", req.Action)
	}
}

func runErrorsOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Action {
	case "missing-args":
		sendResp, err := execAgentRun(t, req, req.SendArgs...)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		return resp, nil
	case "session-not-found":
		sendResp, err := execSendToSession(t, req, "bogus-session-id", req.SendMessage)
		if err != nil {
			return resp, err
		}
		copySendResp(resp, sendResp)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown errors action %q", req.Action)
	}
}

func runAliasOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	startIdleStubSession(t, req)
	shortcut, err := execAgentRun(t, req, "send", req.TerminalSessionID, req.SendMessage)
	if err != nil {
		return resp, err
	}
	ttySend, err := execAgentRun(t, req, "tty", "send", req.TerminalSessionID, "via-tty-send")
	if err != nil {
		return resp, err
	}
	resp.ShortcutStdout = shortcut.Stdout
	resp.ShortcutMsgID = strings.TrimSpace(shortcut.Stdout)
	resp.TTYSubcmdStdout = ttySend.Stdout
	resp.TTYSubcmdMsgID = strings.TrimSpace(ttySend.Stdout)
	resp.ExitCode = shortcut.ExitCode
	resp.Stderr = shortcut.Stderr
	if ttySend.ExitCode != 0 {
		resp.ExitCode = ttySend.ExitCode
		resp.Stderr = ttySend.Stderr
	}
	return resp, nil
}

func copySendResp(dst *Response, src *cliResponse) {
	dst.Stdout = src.Stdout
	dst.Stderr = src.Stderr
	dst.ExitCode = src.ExitCode
	dst.Err = src.Err
}

type cliResponse struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func execSend(t *testing.T, req *Request, message string, extraFlags ...string) (*cliResponse, error) {
	t.Helper()
	args := buildSendArgs(req, message, extraFlags...)
	return execAgentRun(t, req, args...)
}

func execSendToSession(t *testing.T, req *Request, sessionID, message string, extraFlags ...string) (*cliResponse, error) {
	t.Helper()
	old := req.TerminalSessionID
	req.TerminalSessionID = sessionID
	defer func() { req.TerminalSessionID = old }()
	return execSend(t, req, message, extraFlags...)
}

func buildSendArgs(req *Request, message string, extraFlags ...string) []string {
	if len(req.SendArgs) > 0 {
		return req.SendArgs
	}
	var args []string
	if req.UseTTYSubcmd {
		args = []string{"tty", "send"}
	} else {
		args = []string{"send"}
	}
	args = append(args, extraFlags...)
	if req.TerminalSessionID != "" {
		args = append(args, req.TerminalSessionID)
	}
	if message != "" {
		args = append(args, message)
	}
	return args
}

func sessionMessageRef(sessionID, msgID string) string {
	return sessionID + "/" + msgID
}

func execAgentRun(t *testing.T, req *Request, args ...string) (*cliResponse, error) {
	t.Helper()
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &cliResponse{
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

func execSendCaptureFirstStdoutLine(t *testing.T, req *Request, timeout time.Duration, message string, extraFlags ...string) (idLine string, latency time.Duration, resp *cliResponse, err error) {
	t.Helper()
	args := buildSendArgs(req, message, extraFlags...)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", 0, nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return "", 0, nil, err
	}
	reader := bufio.NewReader(stdoutPipe)
	idLine, readErr := reader.ReadString('\n')
	latency = time.Since(start)
	if readErr != nil && readErr != io.EOF {
		_ = cmd.Process.Kill()
		return "", latency, nil, readErr
	}
	idLine = strings.TrimSpace(idLine)
	waitErr := cmd.Wait()
	cli := &cliResponse{Stdout: idLine + "\n", Stderr: stderr.String(), Err: waitErr}
	if waitErr == nil {
		return idLine, latency, cli, nil
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		cli.ExitCode = exitErr.ExitCode()
		return idLine, latency, cli, nil
	}
	if ctx.Err() != nil {
		return idLine, latency, cli, ctx.Err()
	}
	return idLine, latency, cli, waitErr
}

func queueFilePath(home, runner, sessionID string) string {
	return filepath.Join(home, "send-queue", runner, sessionID+".jsonl")
}

func queueContainsMsgID(t *testing.T, path, msgID string) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		t.Fatalf("read queue file: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if id, _ := entry["id"].(string); id == msgID {
			return true
		}
	}
	return false
}

func stubScenarioKeepAliveJSON() string {
	return `{
  "banner_delay_ms": 30,
  "banner_text": "STUB_TTY_BANNER ready",
  "screen_status": "idle",
  "screen_frames": [
    {"delay_ms": 0, "text": "STUB_TTY_BANNER\n› "}
  ],
  "llm_events": [],
  "runner_session_id": "stub-session-keep",
  "exit_after_turn": false
}`
}

func stubScenarioBusyJSON() string {
	return `{
  "banner_delay_ms": 0,
  "banner_text": "STUB_TTY_BANNER",
  "screen_status": "busy",
  "screen_frames": [
    {"delay_ms": 0, "text": "STUB_TTY_BANNER\n• Working on task...\n"}
  ],
  "writable_reason": "stub waiting for turn complete",
  "exit_after_turn": false
}`
}

func stubScenarioBusyThenIdleJSON() string {
	return `{
  "banner_delay_ms": 0,
  "banner_text": "STUB_TTY_BANNER",
  "screen_status": "busy",
  "screen_frames": [
    {"delay_ms": 0, "text": "STUB_TTY_BANNER\n• Working on task...\n"},
    {"delay_ms": 12000, "text": "STUB_TTY_BANNER\n› "}
  ],
  "writable_reason": "stub waiting for turn complete",
  "exit_after_turn": false
}`
}

func writeStubScenario(t *testing.T, req *Request) string {
	t.Helper()
	content := req.StubScenarioJSON
	if content == "" {
		content = stubScenarioKeepAliveJSON()
	}
	path := req.StubScenarioPath
	if path == "" {
		path = filepath.Join(req.TempDir, "stub-scenario.json")
		req.StubScenarioPath = path
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write stub scenario: %v", err)
	}
	return path
}

func stubEnvWithScenario(req *Request, scenarioPath string) []string {
	env := append([]string{}, req.Env...)
	env = append(env, "AGENT_RUN_ENABLE_STUB_TTY=1")
	env = append(env, "AGENT_RUN_STUB_TTY_SCENARIO="+scenarioPath)
	return env
}

func startIdleStubSession(t *testing.T, req *Request) {
	t.Helper()
	req.EnableStubTTY = true
	if req.StubScenarioJSON == "" {
		req.StubScenarioJSON = stubScenarioKeepAliveJSON()
	}
	req.TerminalSessionID = startStubTTYBackground(t, req)
}

func startBusyStubSession(t *testing.T, req *Request) {
	t.Helper()
	req.EnableStubTTY = true
	req.StubScenarioJSON = stubScenarioBusyJSON()
	req.TerminalSessionID = startStubTTYBackground(t, req)
}

func startStubTTYBackground(t *testing.T, req *Request) string {
	t.Helper()
	scenarioPath := writeStubScenario(t, req)
	args := []string{"run", "--agent-runner", "stub-tty", "--keep-tty", "send-queue-probe"}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), stubEnvWithScenario(req, scenarioPath)...)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start stub-tty background: %v", err)
	}
	req.BackgroundCmd = cmd
	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
	sessionID := waitForStubSessionLine(t, stderr, 45*time.Second)
	if sessionID == "" {
		t.Fatalf("stub-tty session id not found in stderr:\n%s", stderr.String())
	}
	addr := readRegistryListenAddr(t, req.Home, req.RunnerID, sessionID)
	waitForPortOpen(t, addr, 10*time.Second)
	return sessionID
}

func waitForStubSessionLine(t *testing.T, r io.Reader, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 4096)
	acc := ""
	for time.Now().Before(deadline) {
		n, err := r.Read(buf)
		if n > 0 {
			acc += string(buf[:n])
			if m := stubSessionIDRe.FindStringSubmatch(acc); len(m) > 1 {
				return m[1]
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			return ""
		}
	}
	if m := stubSessionIDRe.FindStringSubmatch(acc); len(m) > 1 {
		return m[1]
	}
	return ""
}

func readRegistryListenAddr(t *testing.T, home, runner, sessionID string) string {
	t.Helper()
	path := filepath.Join(home, runner+"-registry", sessionID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("parse registry: %v", err)
	}
	addr, _ := entry["listen_addr"].(string)
	return addr
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

var injectionMarkers = []string{"fifo-message-A", "fifo-message-B", "hello", "first-message", "second-message", "writable-wait-probe", "max-wait-probe", "no-wait-probe", "cancel-probe", "fifo-trigger", "via-tty-send", "alias-probe", "delivered-probe", "timeout-probe"}

type injectionCapture struct {
	t             *testing.T
	output        strings.Builder
	mu            sync.Mutex
	seen          []string
	stopCh        chan struct{}
	done          chan struct{}
	agentRun      string
	env           []string
	sessionID     string
}

func injectionCaptureRecordText(c *injectionCapture, text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.output.WriteString(text)
	for _, marker := range injectionMarkers {
		if strings.Contains(text, marker) && !containsString(c.seen, marker) {
			c.seen = append(c.seen, marker)
		}
	}
}

func injectionCapturePollSnapshot(c *injectionCapture) {
	if c == nil || c.agentRun == "" || c.sessionID == "" {
		return
	}
	cmd := exec.Command(c.agentRun, "tty", "snapshot", c.sessionID)
	cmd.Env = append(os.Environ(), c.env...)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	injectionCaptureRecordText(c, string(out))
}

func startInjectionCapture(t *testing.T, req *Request) *injectionCapture {
	t.Helper()
	addr := readRegistryListenAddr(t, req.Home, req.RunnerID, req.TerminalSessionID)
	u, err := url.Parse("http://" + addr)
	if err != nil {
		t.Fatalf("parse listen addr: %v", err)
	}
	u.Scheme = "ws"
	u.Path = "/api/terminal"
	q := u.Query()
	q.Set("session_id", req.TerminalSessionID)
	q.Set("attach_mode", "observer")
	u.RawQuery = q.Encode()
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial observer ws: %v", err)
	}
	cap := &injectionCapture{
		t:         t,
		stopCh:    make(chan struct{}),
		done:      make(chan struct{}),
		agentRun:  req.AgentRun,
		env:       append([]string{}, req.Env...),
		sessionID: req.TerminalSessionID,
	}
	go func() {
		defer close(cap.done)
		defer conn.Close()
		defer func() { recover() }()
		for {
			select {
			case <-cap.stopCh:
				return
			default:
			}
			_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			_, msg, err := conn.ReadMessage()
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}
				return
			}
			injectionCaptureRecordText(cap, string(msg))
		}
	}()
	time.Sleep(200 * time.Millisecond)
	return cap
}

func injectionCaptureSnapshot(c *injectionCapture) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := append([]string{}, c.seen...)
	return out
}

func injectionCaptureWaitFor(c *injectionCapture, marker string, timeout time.Duration) []string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		injectionCapturePollSnapshot(c)
		if containsString(injectionCaptureSnapshot(c), marker) {
			return injectionCaptureSnapshot(c)
		}
		time.Sleep(100 * time.Millisecond)
	}
	c.t.Fatalf("timeout waiting for injection marker %q; seen=%v output tail:\n%s", marker, injectionCaptureSnapshot(c), tailString(c.output.String(), 500))
	return nil
}

func injectionCaptureWaitForOrdered(c *injectionCapture, markers []string, timeout time.Duration) []string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		injectionCapturePollSnapshot(c)
		seen := injectionCaptureSnapshot(c)
		if orderedSubsequence(seen, markers) {
			return filterOrdered(seen, markers)
		}
		time.Sleep(100 * time.Millisecond)
	}
	c.t.Fatalf("timeout waiting for ordered injection %v; seen=%v", markers, injectionCaptureSnapshot(c))
	return nil
}

func injectionCaptureStop(c *injectionCapture) {
	close(c.stopCh)
	<-c.done
}

func containsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func orderedSubsequence(seen, want []string) bool {
	if len(want) == 0 {
		return true
	}
	i := 0
	for _, s := range seen {
		if s == want[i] {
			i++
			if i == len(want) {
				return true
			}
		}
	}
	return false
}

func filterOrdered(seen, want []string) []string {
	var out []string
	i := 0
	for _, s := range seen {
		if i < len(want) && s == want[i] {
			out = append(out, s)
			i++
		}
	}
	return out
}

func tailString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
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

func assertMsgIDLine(t *testing.T, got string) {
	t.Helper()
	line := strings.TrimSpace(got)
	if !msgIDLineRe.MatchString(line) {
		t.Fatalf("expected msg_<n> stdout line, got %q", got)
	}
}
```