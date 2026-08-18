package termtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
)

// Request is the doctest harness request for agent-term CLI tests.
type Request struct {
	Phase string

	Args       []string
	ListenAddr string
	DaemonPort int

	StartDaemon        bool
	EnsureNoDaemon     bool
	RenameBeforeAttach string
	WebSessionName     string
	RunCommand         []string
	DetachSignal       bool
	RunProbeSeconds    int
	RequireGrok        bool

	AgentTermBin string
	DaemonAddr   string
}

// Response is the doctest harness response for agent-term CLI tests.
type Response struct {
	ExitCode   int
	Stdout     string
	Stderr     string
	Combined   string
	HTTPStatus int
	HTTPBody   string
	DaemonPort int

	DaemonStderr          string
	SessionStillRunning   bool
	DetachedSessionID     string
}

// Run executes an agent-term doctest phase.
func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Phase {
	case "serve-tcp":
		return runServeAcceptsTCP(t, req)
	case "list-no-daemon":
		return runCLI(t, req)
	case "run-exit-id":
		return runRunPrintsID(t, req)
	case "wait-session":
		return runWaitSession(t, req)
	case "attach-by-name":
		return runAttachByName(t, req)
	case "web-xterm":
		return runWebServesPage(t, req)
	case "serve-logs-startup":
		return runServeLogsStartup(t, req)
	case "serve-logs-on-create":
		return runServeLogsOnCreate(t, req)
	case "run-pty":
		return runWithPTY(t, req)
	case "run-pty-probe":
		return runWithPTYProbe(t, req)
	case "run-requires-tty":
		return runRequiresTTY(t, req)
	case "run-detach-survives":
		return runDetachSurvives(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

var (
	cachedAgentTermBin     string
	cachedAgentTermBinErr  error
	cachedAgentTermBinOnce sync.Once
)

// BuildAgentTerm builds the agent-term binary for doctest harnesses.
// start is typically d.DOCTEST_ROOT so the walk can find the repo-root go.mod
// even when the generated test's cwd is a temp directory.
func BuildAgentTerm(t *testing.T, start string) string {
	t.Helper()
	cachedAgentTermBinOnce.Do(func() {
		root, err := findModuleRoot(start)
		if err != nil {
			cachedAgentTermBinErr = err
			return
		}
		out := filepath.Join(os.TempDir(), "agent-term-doctest-shared")
		// Nested cmd module (cmd/go.mod): build from that module via -C.
		cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", out, "./agent-term")
		cmd.Dir = root
		if combined, err := cmd.CombinedOutput(); err != nil {
			cachedAgentTermBinErr = fmt.Errorf("build agent-term: %v\n%s", err, combined)
			return
		}
		cachedAgentTermBin = out
	})
	if cachedAgentTermBinErr != nil {
		t.Fatal(cachedAgentTermBinErr)
	}
	return cachedAgentTermBin
}

func runServeAcceptsTCP(t *testing.T, req *Request) (*Response, error) {
	daemonStartMu.Lock()
	defer daemonStartMu.Unlock()

	resp := &Response{}
	port, err := pickFreePort(37681)
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cmd := exec.Command(req.AgentTermBin, "serve", "--listen", addr)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	t.Cleanup(func() { terminateProcess(cmd) })
	if err := waitTCPReady(addr, 10*time.Second); err != nil {
		return nil, err
	}
	httpResp, err := http.Get(fmt.Sprintf("http://%s/api/terminal/sessions", addr))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	body, _ := io.ReadAll(httpResp.Body)
	resp.DaemonPort = port
	resp.HTTPStatus = httpResp.StatusCode
	resp.HTTPBody = string(body)
	resp.Stderr = stderr.String()
	return resp, nil
}

func runCLI(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	argv := append([]string{}, req.Args...)
	cmd := exec.Command(req.AgentTermBin, argv...)
	if req.EnsureNoDaemon {
		port := req.DaemonPort
		if port <= 0 {
			p, err := pickFreePort(47681)
			if err != nil {
				return nil, err
			}
			port = p
		}
		cmd.Env = append(os.Environ(), fmt.Sprintf("AGENT_TERM_SERVER=http://127.0.0.1:%d", port))
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			return nil, runErr
		}
	}
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.Combined = combineCLIOutput(resp.Stdout, resp.Stderr)
	return resp, nil
}

func combineCLIOutput(stdout, stderr string) string {
	stdout = strings.TrimSpace(stdout)
	switch {
	case stdout != "" && stderr != "":
		return strings.TrimRight(stdout+"\n"+stderr, "\n")
	case stderr != "":
		return strings.TrimRight(stderr, "\n")
	default:
		return stdout
	}
}

func runRunPrintsID(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()
	runArgv := []string{"true"}
	if len(req.RunCommand) > 0 {
		runArgv = req.RunCommand
	}
	env := append(os.Environ(), "AGENT_TERM_SERVER=http://127.0.0.1:"+strconv.Itoa(port))
	cmd := exec.Command(req.AgentTermBin, append([]string{"run"}, runArgv...)...)
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	resp := &Response{Stdout: stdout.String(), Stderr: stderr.String()}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			return nil, runErr
		}
	}
	resp.Combined = combineCLIOutput(resp.Stdout, resp.Stderr)
	return resp, nil
}

func runWaitSession(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()

	command := []string{"sleep", "2"}
	if len(req.RunCommand) > 0 {
		command = req.RunCommand
	}
	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	id, err := createSessionViaAPI(t, base, command, "")
	if err != nil {
		return nil, err
	}

	c := ptyclient.NewClient(base)
	var waitErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				waitErr = fmt.Errorf("panic: %v", r)
			}
		}()
		waitErr = ptyclient.WaitSession(c, id)
	}()
	resp := &Response{Stdout: id}
	if waitErr != nil {
		resp.Stderr = waitErr.Error()
		resp.ExitCode = 1
	}
	return resp, nil
}

func runAttachByName(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()
	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	id, err := createSessionViaAPI(t, base, []string{"sleep", "60"}, req.RenameBeforeAttach)
	if err != nil {
		return nil, err
	}
	if err := renameSessionViaAPI(t, base, id, req.RenameBeforeAttach); err != nil {
		return nil, err
	}
	ok, err := probeAttachByName(base, req.RenameBeforeAttach)
	resp := &Response{}
	if err != nil {
		resp.Stderr = err.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	if !ok {
		resp.ExitCode = 1
		resp.Stderr = "attach probe failed"
	}
	return resp, nil
}

func runWebServesPage(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()
	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	name := req.WebSessionName
	if name == "" {
		name = "web-target"
	}
	id, err := createSessionViaAPI(t, base, []string{"sleep", "120"}, name)
	if err != nil {
		return nil, err
	}
	webPort, err := pickFreePort(38681)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(req.AgentTermBin, "web", id, "--port", strconv.Itoa(webPort))
	cmd.Env = append(os.Environ(), "AGENT_TERM_SERVER="+base)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	t.Cleanup(func() { terminateProcess(cmd) })
	pageURL := fmt.Sprintf("http://127.0.0.1:%d/", webPort)
	if err := waitHTTPReady(pageURL, 30*time.Second); err != nil {
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
		}
		return nil, fmt.Errorf("web server not ready: %w", err)
	}
	httpResp, err := http.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	body, _ := io.ReadAll(httpResp.Body)
	return &Response{HTTPStatus: httpResp.StatusCode, HTTPBody: string(body)}, nil
}

var daemonStartMu sync.Mutex

type daemonHandle struct {
	port    int
	stderr  *bytes.Buffer
	cleanup func()
}

func startDaemon(t *testing.T, req *Request) (int, *daemonHandle, error) {
	daemonStartMu.Lock()
	defer daemonStartMu.Unlock()

	port := req.DaemonPort
	if port <= 0 {
		p, err := pickFreePort(37681)
		if err != nil {
			return 0, nil, err
		}
		port = p
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cmd := exec.Command(req.AgentTermBin, "serve", "--listen", addr)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return 0, nil, err
	}
	if err := waitTCPReady(addr, 15*time.Second); err != nil {
		terminateProcess(cmd)
		return 0, nil, fmt.Errorf("%w (stderr: %s)", err, stderr.String())
	}
	apiURL := fmt.Sprintf("http://%s/api/terminal/sessions", addr)
	if err := waitHTTPReady(apiURL, 15*time.Second); err != nil {
		terminateProcess(cmd)
		return 0, nil, fmt.Errorf("%w (stderr: %s)", err, stderr.String())
	}
	return port, &daemonHandle{
		port:    port,
		stderr:  &stderr,
		cleanup: func() { terminateProcess(cmd) },
	}, nil
}

func runServeLogsStartup(t *testing.T, req *Request) (*Response, error) {
	daemonStartMu.Lock()
	defer daemonStartMu.Unlock()

	port, err := pickFreePort(37681)
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cmd := exec.Command(req.AgentTermBin, "serve", "--listen", addr)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	t.Cleanup(func() { terminateProcess(cmd) })
	if err := waitTCPReady(addr, 10*time.Second); err != nil {
		return nil, fmt.Errorf("%w (stderr: %s)", err, stderr.String())
	}
	time.Sleep(50 * time.Millisecond)
	return &Response{
		DaemonPort:   port,
		DaemonStderr: trimDaemonLog(stderr.String()),
		Stderr:       stderr.String(),
	}, nil
}

func runServeLogsOnCreate(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()
	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	id, err := createSessionViaAPI(t, base, []string{"true"}, "")
	if err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	return &Response{
		DaemonPort:   port,
		DaemonStderr: trimDaemonLog(daemon.stderr.String()),
		HTTPBody:     id,
	}, nil
}

func runWithPTY(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()
	return execPTYRun(t, req.AgentTermBin, port, req)
}

func runWithPTYProbe(t *testing.T, req *Request) (*Response, error) {
	if req.RequireGrok {
		if _, err := exec.LookPath("grok"); err != nil {
			t.Skip("grok not found in PATH")
		}
	}
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()
	return execPTYProbe(t, req.AgentTermBin, port, req)
}

func execPTYRun(t *testing.T, bin string, port int, req *Request) (*Response, error) {
	runArgv := []string{"true"}
	if len(req.RunCommand) > 0 {
		runArgv = req.RunCommand
	}
	env := append(os.Environ(), "AGENT_TERM_SERVER=http://127.0.0.1:"+strconv.Itoa(port))
	cmd := exec.Command(bin, append([]string{"run"}, runArgv...)...)
	cmd.Env = env

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		_, _ = io.Copy(&output, ptmx)
	}()

	if req.DetachSignal {
		time.Sleep(500 * time.Millisecond)
		_ = cmd.Process.Signal(syscall.SIGINT)
	}

	runErr := cmd.Wait()
	_ = ptmx.Close()
	<-readDone

	resp := &Response{
		DaemonPort: port,
		Stdout:     output.String(),
		Combined:   strings.TrimSpace(output.String()),
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			return nil, runErr
		}
	}
	return resp, nil
}

func execPTYProbe(t *testing.T, bin string, port int, req *Request) (*Response, error) {
	probe := req.RunProbeSeconds
	if probe <= 0 {
		probe = 4
	}
	runArgv := []string{"sleep", "60"}
	if len(req.RunCommand) > 0 {
		runArgv = req.RunCommand
	}
	env := append(os.Environ(), "AGENT_TERM_SERVER=http://127.0.0.1:"+strconv.Itoa(port))
	cmd := exec.Command(bin, append([]string{"run"}, runArgv...)...)
	cmd.Env = env

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		_, _ = io.Copy(&output, ptmx)
	}()

	time.Sleep(time.Duration(probe) * time.Second)
	_ = cmd.Process.Signal(syscall.SIGTERM)
	time.Sleep(200 * time.Millisecond)
	terminateProcess(cmd)
	_ = ptmx.Close()
	<-readDone

	text := output.String()
	resp := &Response{
		DaemonPort: port,
		Stdout:     text,
		Combined:   strings.TrimSpace(text),
		ExitCode:   1,
	}
	return resp, nil
}

func runRequiresTTY(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()

	runArgv := []string{"bash"}
	if len(req.RunCommand) > 0 {
		runArgv = req.RunCommand
	}
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		stdinR.Close()
		stdinW.Close()
		return nil, err
	}
	t.Cleanup(func() {
		stdinR.Close()
		stdinW.Close()
		stdoutR.Close()
		stdoutW.Close()
	})

	env := append(os.Environ(), "AGENT_TERM_SERVER=http://127.0.0.1:"+strconv.Itoa(port))
	cmd := exec.Command(req.AgentTermBin, append([]string{"run"}, runArgv...)...)
	cmd.Env = env
	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	_ = stdinW.Close()
	_ = stdoutW.Close()

	go func() { _, _ = io.Copy(io.Discard, stdoutR) }()

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	resp := &Response{DaemonPort: port}
	select {
	case runErr := <-done:
		if runErr != nil {
			if exitErr, ok := runErr.(*exec.ExitError); ok {
				resp.ExitCode = exitErr.ExitCode()
			} else {
				return nil, runErr
			}
		}
		resp.Stderr = strings.TrimSpace(stderr.String())
		resp.Combined = strings.TrimRight(stderr.String(), "\n")
	case <-time.After(8 * time.Second):
		terminateProcess(cmd)
		resp.ExitCode = 124
		resp.Stderr = "timeout waiting for TTY requirement error"
		resp.Combined = resp.Stderr
	}
	return resp, nil
}

func runDetachSurvives(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()

	detachReq := *req
	detachReq.RunCommand = []string{"sleep", "60"}
	detachReq.DetachSignal = true
	resp, err := execPTYRun(t, req.AgentTermBin, port, &detachReq)
	if err != nil {
		return nil, err
	}

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	c := ptyclient.NewClient(base)
	sessions, err := c.List()
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(sessions)
	resp.HTTPBody = string(data)
	for _, s := range sessions {
		if s.Status == "running" {
			resp.SessionStillRunning = true
			resp.DetachedSessionID = s.ID
			break
		}
	}
	return resp, nil
}

func trimDaemonLog(stderr string) string {
	return strings.TrimRight(stderr, "\n")
}

func findModuleRoot(start string) (string, error) {
	candidates := make([]string, 0, 3)
	if start != "" {
		candidates = append(candidates, start)
	}
	if root := os.Getenv("DOCTEST_ROOT"); root != "" {
		candidates = append(candidates, root)
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd)
	}
	seen := map[string]bool{}
	for _, startDir := range candidates {
		for dir := startDir; ; dir = filepath.Dir(dir) {
			if !seen[dir] {
				seen[dir] = true
				if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
					if _, err := os.Stat(filepath.Join(dir, "cmd/agent-term")); err == nil {
						return dir, nil
					}
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	return "", fmt.Errorf("go.mod not found")
}

// PickFreePort finds an available TCP port near base.
func PickFreePort(base int) (int, error) {
	return pickFreePort(base)
}

func pickFreePort(base int) (int, error) {
	_ = base
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	p := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return p, nil
}

func waitTCPReady(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for tcp %s", addr)
}

func waitHTTPReady(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}

func terminateProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	cmd.Process.Signal(syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
	cmd.Process.Kill()
}

func createSessionViaAPI(t *testing.T, base string, command []string, name string) (string, error) {
	t.Helper()
	body := map[string]interface{}{"name": name}
	if len(command) > 0 {
		body["command"] = command[0]
		if len(command) > 1 {
			body["args"] = command[1:]
		}
	}
	data, _ := json.Marshal(body)
	resp, err := http.Post(base+"/api/terminal/sessions", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", fmt.Errorf("empty id from create")
	}
	return out.ID, nil
}

func renameSessionViaAPI(t *testing.T, base, id, name string) error {
	req, err := http.NewRequest(http.MethodPatch, base+"/api/terminal/sessions/"+id,
		strings.NewReader(fmt.Sprintf(`{"name":%q}`, name)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rename status %d", resp.StatusCode)
	}
	return nil
}

func probeAttachByName(base, name string) (bool, error) {
	c := ptyclient.NewClient(base)
	session, err := ptyclient.ResolveTarget(c, name)
	if err != nil {
		return false, err
	}
	u, err := url.Parse(base)
	if err != nil {
		return false, err
	}
	u.Scheme = "ws"
	u.Path = "/api/terminal"
	q := u.Query()
	q.Set("session_id", session.ID)
	u.RawQuery = q.Encode()
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return false, err
		}
		var m struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
		}
		if json.Unmarshal(msg, &m) == nil && m.Type == "session_id" && m.SessionID == session.ID {
			return true, nil
		}
	}
	return false, fmt.Errorf("timeout waiting for session_id")
}