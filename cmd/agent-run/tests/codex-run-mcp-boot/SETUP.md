# Scenario

**Feature**: `agent-run run --detach` + `llm-mock-run-codex` with MCP servers in isolated CODEX_HOME

```
build agent-run + llm-mock + llm-mock-run-codex
  -> isolated AGENT_RUN_HOME + CODEX_HOME + workspace
  -> hang MCP scripts + LLM_MOCK_EXTRA_MCP_TOML_FILE
  -> run --detach (empty prompt)
  -> poll tty snapshot until "Starting MCP servers"
  -> classify BannerDetected + CheckWritable
```

## Preconditions

- Nested DOCTEST root. Repo root: `d.DOCTEST_ROOT/../../../..`.
- Real `codex` on PATH (else `t.Skip`).
- Build: `agent-run` from `cmd/`; `llm-mock` + `llm-mock-run-codex` from repo root
  (same cache pattern as `codex-open-bind-resume`).

## Steps

1. Root Setup: homes, binaries, hang MCP toml, skip without codex.
2. Run: detach empty prompt; poll snapshot; classify; kill serve.
3. Assert: MCP boot frame is not inject-ready.

## Context

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/doctest/session"
)

const (
	envLLMMockCodexHome      = "LLM_MOCK_CODEX_HOME"
	envLLMMockConfigFile     = "LLM_MOCK_CONFIG_FILE"
	envLLMMockExtraMCPFile   = "LLM_MOCK_EXTRA_MCP_TOML_FILE"
	defaultSessionID         = "codex-run-mcp-boot-s1"
	defaultExecTimeout       = 90 * time.Second
	defaultMCPPoll           = 45 * time.Second
	mcpServerCount           = 8
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "codex-run-mcp-boot-doctest-"+d.DOCTEST_SESSION_ID)
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
	_, err := os.Stat(path)
	return err == nil
}

func ensureSessionBinaries(t *testing.T, d *session.Doctest, repoRoot string) (agentRun, llmMock, llmMockRunCodex string) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	llmMock = filepath.Join(cache, "llm-mock")
	llmMockRunCodex = filepath.Join(cache, "llm-mock-run-codex")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err := withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(llmMock) && fileExists(llmMockRunCodex) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		cmdDir := filepath.Join(repoRoot, "cmd")
		b1 := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", agentRun, "./agent-run")
		b1.Dir = cmdDir
		if out, err := b1.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, out)
		}
		for _, b := range []struct {
			out string
			pkg string
		}{
			{llmMock, "./agent/llm/llm-mock"},
			{llmMockRunCodex, "./agent/llm/llm-mock/llm-mock-run-codex"},
		} {
			cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", b.out, b.pkg)
			cmd.Dir = repoRoot
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("build %s: %w\n%s", b.pkg, err, out)
			}
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	if err != nil {
		t.Fatalf("session binaries: %v", err)
	}
	return agentRun, llmMock, llmMockRunCodex
}

func writeHangMCPAndToml(t *testing.T, dir string) (script, extra string) {
	t.Helper()
	script = filepath.Join(dir, "hang-mcp.sh")
	body := "#!/bin/sh\nexec sleep 600\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatalf("write hang-mcp: %v", err)
	}
	var b strings.Builder
	for i := 1; i <= mcpServerCount; i++ {
		fmt.Fprintf(&b, "[mcp_servers.slowinit_%02d]\ncommand = %q\nargs = [%q]\n\n", i, script, fmt.Sprintf("%d", i))
	}
	extra = filepath.Join(dir, "extra-mcp.toml")
	if err := os.WriteFile(extra, []byte(b.String()), 0644); err != nil {
		t.Fatalf("write extra-mcp.toml: %v", err)
	}
	return script, extra
}

func withoutEnvKey(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func baseEnv(req *Request) []string {
	env := os.Environ()
	binDir := filepath.Dir(req.LLMMockRunCodex)
	path := binDir + string(os.PathListSeparator) + os.Getenv("PATH")
	env = withoutEnvKey(env, "PATH")
	env = append(env, "PATH="+path)
	env = withoutEnvKey(env, "AGENT_RUN_HOME")
	env = append(env, "AGENT_RUN_HOME="+req.Home)
	env = withoutEnvKey(env, envLLMMockConfigFile)
	env = append(env, envLLMMockConfigFile+"="+req.MockConfigFile)
	env = withoutEnvKey(env, envLLMMockCodexHome)
	env = append(env, envLLMMockCodexHome+"="+req.CodexHome)
	env = withoutEnvKey(env, envLLMMockExtraMCPFile)
	env = append(env, envLLMMockExtraMCPFile+"="+req.ExtraMCPFile)
	env = withoutEnvKey(env, "CODEX_HOME")
	env = append(env, "CODEX_HOME="+req.CodexHome)
	env = withoutEnvKey(env, "AGENT_RUN_CODEX_TTY_COMMAND")
	return env
}

func cleanupRegistryServes(home string) {
	reg := filepath.Join(home, "codex-tty-registry")
	entries, err := os.ReadDir(reg)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(reg, e.Name()))
		if err != nil {
			continue
		}
		var st struct {
			PID int `json:"pid"`
		}
		if json.Unmarshal(data, &st) != nil || st.PID <= 0 {
			continue
		}
		_ = exec.Command("pkill", "-P", fmt.Sprintf("%d", st.PID)).Run()
		if p, err := os.FindProcess(st.PID); err == nil {
			_ = p.Signal(syscall.SIGTERM)
			time.Sleep(100 * time.Millisecond)
			_ = p.Signal(syscall.SIGKILL)
		}
	}
}

func execAgentRun(t *testing.T, req *Request, timeout time.Duration, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	if timeout <= 0 {
		timeout = req.ExecTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.Workspace
	cmd.Env = baseEnv(req)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func runDetachAndCaptureMCPBoot(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })

	args := []string{
		"run",
		"--detach",
		"--agent-runner=codex-tty",
		"--agent-runner-binary=" + req.LLMMockRunCodex,
		"--agent-runner-config-home=" + req.CodexHome,
		"--session-id=" + req.SessionID,
		"--dir=" + req.Workspace,
		"--prepend-path=" + filepath.Dir(req.LLMMockRunCodex),
		"--env", envLLMMockCodexHome + "=" + req.CodexHome,
		"--env", envLLMMockExtraMCPFile + "=" + req.ExtraMCPFile,
		"--env", envLLMMockConfigFile + "=" + req.MockConfigFile,
	}
	ctx, cancel := context.WithTimeout(context.Background(), req.ExecTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.Workspace
	cmd.Env = baseEnv(req)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start detach: %w", err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
	})

	resp := &Response{}
	poll := req.MCPPoll
	if poll <= 0 {
		poll = defaultMCPPoll
	}
	deadline := time.Now().Add(poll)
	var lastSnap string
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(filepath.Join(req.CodexHome, "config.toml")); err == nil {
			resp.ConfigTOML = string(data)
		}
		snapOut, _, err := execAgentRun(t, req, 5*time.Second, "tty", "snapshot", req.SessionID)
		if err == nil {
			lastSnap = snapOut
			if strings.Contains(strings.ToLower(snapOut), "starting mcp") {
				resp.Snapshot = snapOut
				resp.SawMCPBoot = true
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	resp.DetachStderr = stderr.String()
	if resp.Snapshot == "" {
		resp.Snapshot = lastSnap
	}
	if resp.ConfigTOML == "" {
		if data, err := os.ReadFile(filepath.Join(req.CodexHome, "config.toml")); err == nil {
			resp.ConfigTOML = string(data)
		}
	}
	if resp.Snapshot != "" {
		p, ok := agenttty.Get("codex-tty")
		if !ok {
			return resp, fmt.Errorf("codex-tty provider missing")
		}
		resp.Writable = p.CheckWritable([]byte(resp.Snapshot))
		resp.BannerReady = agenttty.BannerDetected([]byte(resp.Snapshot), "codex", []string{"CODEX_TTY_BANNER"})
		if dump := strings.TrimSpace(req.SnapshotDumpPath); dump != "" {
			if err := os.MkdirAll(filepath.Dir(dump), 0755); err != nil {
				return resp, fmt.Errorf("mkdir snapshot dump: %w", err)
			}
			if err := os.WriteFile(dump, []byte(resp.Snapshot), 0644); err != nil {
				return resp, fmt.Errorf("write snapshot dump: %w", err)
			}
		}
	}
	return resp, nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not found in PATH")
	}
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.CodexHome = filepath.Join(req.TempDir, ".codex")
	req.Workspace = filepath.Join(req.TempDir, "workspace")
	for _, dir := range []string{req.Home, req.CodexHome, req.Workspace} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	_ = exec.Command("git", "init", "-q", req.Workspace).Run()
	if err := os.WriteFile(filepath.Join(req.Workspace, "README.md"), []byte("codex-run-mcp-boot\n"), 0644); err != nil {
		return err
	}

	req.AgentRun, req.LLMMock, req.LLMMockRunCodex = ensureSessionBinaries(t, d, req.RepoRoot)
	if filepath.Dir(req.LLMMock) != filepath.Dir(req.LLMMockRunCodex) {
		return fmt.Errorf("llm-mock and llm-mock-run-codex must share a directory")
	}

	req.MockConfigFile = filepath.Join(req.TempDir, "llm-mock-config.json")
	if err := os.WriteFile(req.MockConfigFile, []byte(`{"exchanges":[]}`), 0644); err != nil {
		return err
	}
	_, req.ExtraMCPFile = writeHangMCPAndToml(t, req.TempDir)

	if req.SessionID == "" {
		req.SessionID = defaultSessionID
	}
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = defaultExecTimeout
	}
	if req.MCPPoll <= 0 {
		req.MCPPoll = defaultMCPPoll
	}
	if req.SnapshotDumpPath == "" {
		req.SnapshotDumpPath = strings.TrimSpace(os.Getenv("CODEX_MCP_BOOT_DUMP_SNAPSHOT"))
	}
	return nil
}
```
