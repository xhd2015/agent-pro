# Scenario

**Feature**: production open attach (`attach_mode=screen` under OpenCloseExits)
must preserve child mouse-tracking CSI on the host TTY stream

```
# crime-scene stack
build agent-run + llm-mock-run-grok + llm-mock (sibling) + attachhost
agent-run run --keep-tty --agent-runner grok-tty \
  --agent-runner-binary llm-mock-run-grok \
  --agent-runner-config-home <GROK_HOME> \
  --dir <ws> "probe"
  + AGENT_RUN_GROK_TTY_COMMAND unset
  + LLM_MOCK_RUN_GROK_COMMAND unset  # real grok under mock API
  -> registry session + listen_addr

AttachWriter(listen, id, "screen") under host PTY
  -> host bytes must include mouse CSI (desired)
```

## Preconditions

- Nested DOCTEST root; repo root `d.DOCTEST_ROOT/../../../..`.
- Session cache: `$TMPDIR/open-attach-fidelity-doctest-<session>/`.
- **Requires real `grok` on PATH** (skip otherwise).
- Sibling `llm-mock` next to `llm-mock-run-grok` (server discovery).
- No `AGENT_RUN_GROK_TTY_COMMAND` / `LLM_MOCK_RUN_GROK_COMMAND` hooks.

## Steps

1. Root Setup builds binaries, allocates AGENT_RUN_HOME + GROK_HOME + workspace.
2. Leaf sets HostAttachMode=screen.
3. Run starts keep-tty with llm-mock-run-grok, waits registry + TUI settle,
   captures host attach (`screen`) and control attach (`attach`).
4. Assert desired host mouse CSI (RED until product fix).

```go
import (
	"runtime"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/xhd2015/doctest/session"
)

const (
	defaultRunner = "grok-tty"
	defaultPrompt = "open-attach-fidelity-probe"
	envGrokTTYCmd = "AGENT_RUN_GROK_TTY_COMMAND"
	envLLMMockCmd = "LLM_MOCK_RUN_GROK_COMMAND"
)

var (
	CsiAltScreen     = []byte("\x1b[?1049h")
	CsiMouse1000     = []byte("\x1b[?1000h")
	CsiMouse1002     = []byte("\x1b[?1002h")
	CsiMouse1006     = []byte("\x1b[?1006h")
	grokTTYSessionRe = regexp.MustCompile(`(?m)grok-tty:\s*(\S+)`)
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "open-attach-fidelity-doctest-"+d.DOCTEST_SESSION_ID)
}

func withFileLock(t *testing.T, lockPath string, fn func() error) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}

func ensureStubDist(distDir string) error {
	entries, statErr := os.ReadDir(distDir)
	if statErr == nil {
		for _, e := range entries {
			if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				return nil
			}
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<!doctype html><title>stub</title>\n"), 0644)
}

// ensureSessionBinaries builds agent-run, llm-mock-run-grok, sibling llm-mock, attachhost.
func ensureSessionBinaries(t *testing.T, d *session.Doctest, repoRoot string) (agentRun, llmMockRunGrok, attachHelper string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	llmMockRunGrok = filepath.Join(cache, "llm-mock-run-grok")
	llmMock := filepath.Join(cache, "llm-mock")
	attachHelper = filepath.Join(cache, "attachhost")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(llmMockRunGrok) && fileExists(llmMock) && fileExists(attachHelper) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		for _, rel := range []string{"frontend-agent-run/dist", "frontend/dist"} {
			if err := ensureStubDist(filepath.Join(repoRoot, rel)); err != nil {
				return fmt.Errorf("ensure %s stub: %w", rel, err)
			}
		}
		// agent-run lives under nested cmd module
		cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", agentRun, "./agent-run")
		cmd.Dir = filepath.Join(repoRoot, "cmd")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go build agent-run: %w\n%s", err, out)
		}
		// llm-mock-run-grok + sibling llm-mock from agent-pro module
		for _, b := range []struct {
			out  string
			pkg  string
		}{
			{llmMockRunGrok, "./agent/llm/llm-mock/llm-mock-run-grok"},
			{llmMock, "./agent/llm/llm-mock"},
			{attachHelper, "./script/debug/open-attach-fidelity-attachhost"},
		} {
			c := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", b.out, b.pkg)
			c.Dir = repoRoot
			if out, err := c.CombinedOutput(); err != nil {
				return fmt.Errorf("go build %s: %w\n%s", b.pkg, err, out)
			}
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	return agentRun, llmMockRunGrok, attachHelper, err
}

func parseGrokTTYSessionID(stderr string) (string, bool) {
	m := grokTTYSessionRe.FindStringSubmatch(stderr)
	if len(m) < 2 {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

type registryEntry struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
}

func readRegistry(home, sessionID string) (registryEntry, error) {
	path := filepath.Join(home, "grok-tty-registry", sessionID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return registryEntry{}, err
	}
	var e registryEntry
	if err := json.Unmarshal(data, &e); err != nil {
		return registryEntry{}, err
	}
	return e, nil
}

func waitRegistry(home, sessionID string, timeout time.Duration) (registryEntry, error) {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		e, err := readRegistry(home, sessionID)
		if err == nil && strings.TrimSpace(e.ListenAddr) != "" {
			return e, nil
		}
		last = err
		time.Sleep(150 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timeout waiting for registry %s", sessionID)
	}
	return registryEntry{}, last
}

func captureAttachHost(t *testing.T, helper, listen, sessionID, attachMode string, secs float64) ([]byte, error) {
	t.Helper()
	cmd := exec.Command(helper, listen, sessionID, attachMode)
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 30, Cols: 100})
	if err != nil {
		return nil, fmt.Errorf("pty start attachhost: %w", err)
	}
	defer ptmx.Close()

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, ptmx)
		close(done)
	}()

	time.Sleep(time.Duration(secs * float64(time.Second)))
	_ = cmd.Process.Signal(os.Interrupt)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
		<-done
	}
	_ = cmd.Wait()
	return buf.Bytes(), nil
}

func HostHasMouseCSI(b []byte) bool {
	return bytes.Contains(b, CsiMouse1002) || bytes.Contains(b, CsiMouse1006) || bytes.Contains(b, CsiMouse1000)
}

func withoutEnvKeys(env []string, keys ...string) []string {
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k+"="] = true
	}
	out := make([]string, 0, len(env))
	for _, e := range env {
		skip := false
		for p := range drop {
			if strings.HasPrefix(e, p) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, e)
		}
	}
	return out
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("real grok not on PATH (required for llm-mock-run-grok default child)")
	}

	repoRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	req.RepoRoot = repoRoot

	agentRun, llmMock, helper, err := ensureSessionBinaries(t, d, repoRoot)
	if err != nil {
		return err
	}
	req.BinPath = agentRun
	req.LLMMockRunGrok = llmMock
	req.AttachHelper = helper

	tmp := t.TempDir()
	req.Home = filepath.Join(tmp, ".agent-run")
	req.GrokHome = filepath.Join(tmp, ".grok")
	req.Workspace = filepath.Join(tmp, "workspace")
	for _, dir := range []string{req.Home, req.GrokHome, req.Workspace} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	_ = os.WriteFile(filepath.Join(req.Workspace, ".keep"), []byte(""), 0644)

	if req.SessionID == "" {
		req.SessionID = "open-attach-fidelity-1"
	}
	if req.HostAttachMode == "" {
		req.HostAttachMode = "open" // production OpenCloseExits --open path
	}
	return nil
}

func runOpenAttachFidelity(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	// Put llm-mock-run-grok dir first so sibling llm-mock resolves; keep real grok on PATH.
	binDir := filepath.Dir(req.LLMMockRunGrok)
	pathEnv := binDir + string(os.PathListSeparator) + os.Getenv("PATH")

	cmd := exec.Command(req.BinPath,
		"run",
		"--agent-runner", defaultRunner,
		"--agent-runner-binary", req.LLMMockRunGrok,
		"--agent-runner-config-home", req.GrokHome,
		"--keep-tty",
		"--dir", req.Workspace,
		"--session-id", req.SessionID,
		defaultPrompt,
	)
	env := withoutEnvKeys(os.Environ(), envGrokTTYCmd, envLLMMockCmd, "AGENT_RUN_HOME", "GROK_HOME", "PATH")
	env = append(env,
		"AGENT_RUN_HOME="+req.Home,
		"GROK_HOME="+req.GrokHome,
		"PATH="+pathEnv,
	)
	cmd.Env = env

	var stderrBuf bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start keep-tty llm-mock-run-grok: %w", err)
	}
	bgCmd := cmd
	t.Cleanup(func() {
		if bgCmd != nil && bgCmd.Process != nil {
			_ = bgCmd.Process.Kill()
			_, _ = bgCmd.Process.Wait()
		}
	})

	deadline := time.Now().Add(45 * time.Second)
	sessionID := ""
	for time.Now().Before(deadline) {
		if id, ok := parseGrokTTYSessionID(stderrBuf.String()); ok {
			sessionID = id
			break
		}
		if _, err := readRegistry(req.Home, req.SessionID); err == nil {
			sessionID = req.SessionID
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if sessionID == "" {
		sessionID = req.SessionID
	}
	entry, err := waitRegistry(req.Home, sessionID, 30*time.Second)
	if err != nil {
		return &Response{Stderr: stderrBuf.String()}, fmt.Errorf("wait registry: %w\nstderr:\n%s", err, stderrBuf.String())
	}
	req.ListenAddr = entry.ListenAddr
	req.TerminalSessionID = entry.SessionID
	if req.TerminalSessionID == "" {
		req.TerminalSessionID = sessionID
	}

	// Real Grok needs time to enable mouse + paint chrome (crime scene ~4s).
	time.Sleep(4 * time.Second)

	hostBytes, err := captureAttachHost(t, req.AttachHelper, req.ListenAddr, req.TerminalSessionID, req.HostAttachMode, 3.0)
	if err != nil {
		return &Response{Stderr: stderrBuf.String()}, err
	}
	controlBytes, _ := captureAttachHost(t, req.AttachHelper, req.ListenAddr, req.TerminalSessionID, "attach", 2.5)

	return &Response{
		HostBytes:          hostBytes,
		ControlAttachBytes: controlBytes,
		Stderr:             stderrBuf.String(),
	}, nil
}
```
