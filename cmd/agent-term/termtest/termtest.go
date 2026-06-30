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
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

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
	case "attach-by-name":
		return runAttachByName(t, req)
	case "web-xterm":
		return runWebServesPage(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

// BuildAgentTerm builds the agent-term binary for doctest harnesses.
func BuildAgentTerm(t *testing.T) string {
	t.Helper()
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	safe := strings.ReplaceAll(t.Name(), "/", "_")
	out := filepath.Join(os.TempDir(), "agent-term-doctest-"+safe)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/agent-term")
	cmd.Dir = root
	if combined, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build agent-term: %v\n%s", err, combined)
	}
	t.Cleanup(func() { os.Remove(out) })
	return out
}

func runServeAcceptsTCP(t *testing.T, req *Request) (*Response, error) {
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
	resp.Combined = strings.TrimSpace(resp.Stdout + "\n" + resp.Stderr)
	return resp, nil
}

func runRunPrintsID(t *testing.T, req *Request) (*Response, error) {
	port, daemon, err := startDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer daemon.cleanup()
	env := append(os.Environ(), "AGENT_TERM_SERVER=http://127.0.0.1:"+strconv.Itoa(port))
	cmd := exec.Command(req.AgentTermBin, "run", "true")
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
	resp.Combined = strings.TrimSpace(resp.Stdout + "\n" + resp.Stderr)
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

type daemonHandle struct {
	port    int
	cleanup func()
}

func startDaemon(t *testing.T, req *Request) (int, *daemonHandle, error) {
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
		cleanup: func() { terminateProcess(cmd) },
	}, nil
}

func findModuleRoot() (string, error) {
	if root := os.Getenv("DOCTEST_ROOT"); root != "" {
		for dir := root; ; dir = filepath.Dir(dir) {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				if _, err := os.Stat(filepath.Join(dir, "cmd/agent-term")); err == nil {
					return dir, nil
				}
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd/agent-term")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// PickFreePort finds an available TCP port near base.
func PickFreePort(base int) (int, error) {
	return pickFreePort(base)
}

var portSeq atomic.Uint64

func pickFreePort(base int) (int, error) {
	offset := int(portSeq.Add(1) % 200)
	for port := base + offset; port < base+400; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			addr := ln.Addr().(*net.TCPAddr)
			p := addr.Port
			ln.Close()
			return p, nil
		}
	}
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