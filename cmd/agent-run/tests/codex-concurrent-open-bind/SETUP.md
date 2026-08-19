# Scenario

**Feature**: concurrent codex-tty open on shared workspace must bind distinct
prompt-matched `runner_session_id`s (not newest-cwd race).

```
build agent-run + llm-mock + llm-mock-run-codex
  -> isolated AGENT_RUN_HOME + CODEX_HOME + workspace
  -> AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> LLM_MOCK_CONFIG_FILE exchanges for QUESTION_A / QUESTION_B

parallel:
  agent-run run --open --session-id concurrent-A ... -- "QUESTION_A_combined_checkout"
  agent-run run --open --session-id concurrent-B ... -- "QUESTION_B_item_level_refid"

  -> both bound; ids distinct; each rollout's first real user prompt matches
```

## Preconditions

- Nested DOCTEST root (does not inherit `cmd/agent-run/tests` Setup/Run).
- Repo root: `d.DOCTEST_ROOT/../../..`
  (`codex-concurrent-open-bind` → `tests` → `agent-run` → `cmd` → module root…
  wait: tree is under `cmd/agent-run/tests/`, so root is `d.DOCTEST_ROOT/../../../..`).
- Real `codex` on PATH (else `t.Skip`).
- Build:
  - `agent-run` from nested module `cmd/` (`go build ./agent-run`)
  - `llm-mock` + `llm-mock-run-codex` from repo root

## Context

```go
import (
	"runtime"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

const (
	envOpenAttachInstant = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envLLMMockCodexHome  = "LLM_MOCK_CODEX_HOME"
	envLLMMockConfigFile = "LLM_MOCK_CONFIG_FILE"

	defaultSessionIDA = "concurrent-A"
	defaultSessionIDB = "concurrent-B"
	defaultPromptA    = "QUESTION_A_combined_checkout"
	defaultPromptB    = "QUESTION_B_item_level_refid"

	defaultExecTimeout = 2 * time.Minute
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "codex-concurrent-open-bind-doctest-"+d.DOCTEST_SESSION_ID)
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

// ensureStubDist writes a DistComplete SPA stub (index.html with #root + assets/*).
func ensureStubDist(distDir string) error {
	if err := os.MkdirAll(filepath.Join(distDir, "assets"), 0755); err != nil {
		return err
	}
	const shell = `<!doctype html>
<html lang="en">
<head><meta charset="UTF-8"><title>agent-run</title></head>
<body>
<div id="root"></div>
</body>
</html>
`
	indexPath := filepath.Join(distDir, "index.html")
	needIndex := true
	if data, err := os.ReadFile(indexPath); err == nil {
		s := string(data)
		if strings.Contains(s, `id="root"`) || strings.Contains(s, "id='root'") {
			needIndex = false
		}
	}
	if needIndex {
		if err := os.WriteFile(indexPath, []byte(shell), 0644); err != nil {
			return err
		}
	}
	assetPath := filepath.Join(distDir, "assets", "doctest-stub.js")
	if st, err := os.Stat(assetPath); err != nil || st.Size() == 0 {
		if err := os.WriteFile(assetPath, []byte("/* doctest stub */\n"), 0644); err != nil {
			return err
		}
	}
	return nil
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
		// DistComplete stub: index.html with #root + assets/* (placeholder.txt alone is incomplete).
		for _, rel := range []string{"frontend-agent-run/dist", "frontend/dist"} {
			if err := ensureStubDist(filepath.Join(repoRoot, rel)); err != nil {
				return fmt.Errorf("ensure %s stub: %w", rel, err)
			}
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

func writeConcurrentMockConfig(t *testing.T, path string) {
	t.Helper()
	const body = `{
  "exchanges": [
    {
      "request": {"role": "user", "content": "QUESTION_A_combined_checkout", "index": -1},
      "response": {"content": "ANSWER_A_unique_body_for_checkout_split", "finish_reason": "stop"}
    },
    {
      "request": {"role": "user", "content": "QUESTION_B_item_level_refid", "index": -1},
      "response": {"content": "ANSWER_B_unique_body_for_item_refid_map", "finish_reason": "stop"}
    }
  ]
}
`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write mock config: %v", err)
	}
}

func baseEnv(req *Request) []string {
	env := append([]string(nil), req.Env...)
	if len(env) == 0 {
		env = os.Environ()
	}
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
	env = withoutEnvKey(env, "CODEX_HOME")
	env = append(env, "CODEX_HOME="+req.CodexHome)
	env = withoutEnvKey(env, envOpenAttachInstant)
	env = append(env, envOpenAttachInstant+"=1")
	env = withoutEnvKey(env, "AGENT_RUN_CODEX_TTY_COMMAND")
	return env
}

func execAgentRun(t *testing.T, req *Request, timeout time.Duration, args ...string) CmdResult {
	t.Helper()
	if timeout <= 0 {
		timeout = req.ExecTimeout
	}
	if timeout <= 0 {
		timeout = defaultExecTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.Workspace
	cmd.Env = baseEnv(req)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := CmdResult{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
	if err == nil {
		return res
	}
	if ctx.Err() != nil {
		res.Err = ctx.Err()
		res.ExitCode = -1
		return res
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		res.ExitCode = ee.ExitCode()
		res.Err = nil
		return res
	}
	return res
}

func openArgs(req *Request, sessionID, prompt string) []string {
	return []string{
		"run",
		"--agent-runner=codex-tty",
		"--agent-runner-binary=" + req.LLMMockRunCodex,
		"--agent-runner-config-home=" + req.CodexHome,
		"--session-id=" + sessionID,
		"--dir=" + req.Workspace,
		"--open",
		"--",
		prompt,
	}
}

func metaPath(req *Request, sessionID string) string {
	return filepath.Join(req.Home, "sessions", sessionID, "meta.json")
}

func readMeta(req *Request, sessionID string) (map[string]any, string) {
	data, err := os.ReadFile(metaPath(req, sessionID))
	if err != nil {
		return nil, ""
	}
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return nil, ""
	}
	rid, _ := m["runner_session_id"].(string)
	return m, strings.TrimSpace(rid)
}

var rolloutUUIDRe = regexp.MustCompile(`rollout-.*-([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\.jsonl$`)

func listCodexIDs(codexHome string) []string {
	var ids []string
	_ = filepath.Walk(filepath.Join(codexHome, "sessions"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		m := rolloutUUIDRe.FindStringSubmatch(filepath.Base(path))
		if len(m) == 2 {
			ids = append(ids, m[1])
		}
		return nil
	})
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func findRolloutPath(codexHome, sessionID string) string {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ""
	}
	var found string
	_ = filepath.Walk(filepath.Join(codexHome, "sessions"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		m := rolloutUUIDRe.FindStringSubmatch(filepath.Base(path))
		if len(m) == 2 && m[1] == sessionID {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// firstRealUserPrompt skips Codex <environment_context> synthetic user blobs.
func firstRealUserPrompt(rolloutPath string) string {
	if strings.TrimSpace(rolloutPath) == "" {
		return ""
	}
	f, err := os.Open(rolloutPath)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 512*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec struct {
			Type    string `json:"type"`
			Payload struct {
				Type    string `json:"type"`
				Role    string `json:"role"`
				Message string `json:"message"`
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"payload"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		switch rec.Type {
		case "response_item":
			if rec.Payload.Type != "" && rec.Payload.Type != "message" {
				continue
			}
			if strings.TrimSpace(rec.Payload.Role) != "user" {
				continue
			}
			var parts []string
			for _, c := range rec.Payload.Content {
				if t := strings.TrimSpace(c.Text); t != "" {
					parts = append(parts, t)
				}
			}
			text := strings.TrimSpace(strings.Join(parts, "\n"))
			if text == "" || strings.HasPrefix(text, "<environment_context>") {
				continue
			}
			return text
		case "event_msg":
			if rec.Payload.Type != "user_message" {
				continue
			}
			text := strings.TrimSpace(rec.Payload.Message)
			if text == "" || strings.HasPrefix(text, "<environment_context>") {
				continue
			}
			return text
		}
	}
	return ""
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

func wipeSessionState(req *Request) {
	cleanupRegistryServes(req.Home)
	_ = os.RemoveAll(filepath.Join(req.Home, "sessions"))
	_ = os.RemoveAll(filepath.Join(req.Home, "codex-tty-registry"))
	_ = os.RemoveAll(filepath.Join(req.CodexHome, "sessions"))
	// Codex may lock sqlite under CODEX_HOME root; clear common db files best-effort.
	entries, _ := os.ReadDir(req.CodexHome)
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".db") || strings.HasSuffix(name, ".db-wal") || strings.HasSuffix(name, ".db-shm") || name == "sessions" {
			_ = os.RemoveAll(filepath.Join(req.CodexHome, name))
		}
	}
	_ = os.MkdirAll(filepath.Join(req.CodexHome, "sessions"), 0755)
	_ = os.MkdirAll(filepath.Join(req.Home, "sessions"), 0755)
}

func runConcurrentSameWorkspace(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	start := time.Now()
	resp := &Response{PromptByCodexID: map[string]string{}}
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })

	// Real Codex under one CODEX_HOME can flake on concurrent sqlite init.
	// Retry a few times; keep a short stagger so both processes still overlap
	// discovery (crime scene used ~0–2s stagger and still collided on bind).
	const maxAttempts = 3
	var openA, openB CmdResult
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		wipeSessionState(req)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			openA = execAgentRun(t, req, 90*time.Second, openArgs(req, req.SessionIDA, req.PromptA)...)
		}()
		go func() {
			defer wg.Done()
			// 800ms stagger: reduces sqlite init collision while both stay in
			// the openCodexBindBudget / sessionDiscoveryGrace window.
			time.Sleep(800 * time.Millisecond)
			openB = execAgentRun(t, req, 90*time.Second, openArgs(req, req.SessionIDB, req.PromptB)...)
		}()
		wg.Wait()

		// Settle: both metas bound or budget.
		deadline := time.Now().Add(25 * time.Second)
		for time.Now().Before(deadline) {
			_, ra := readMeta(req, req.SessionIDA)
			_, rb := readMeta(req, req.SessionIDB)
			ids := listCodexIDs(req.CodexHome)
			if ra != "" && rb != "" && len(ids) >= 2 {
				break
			}
			time.Sleep(400 * time.Millisecond)
		}
		ids := listCodexIDs(req.CodexHome)
		_, ra := readMeta(req, req.SessionIDA)
		_, rb := readMeta(req, req.SessionIDB)
		if len(ids) >= 2 && ra != "" && rb != "" {
			break
		}
		// Tear down before retry.
		_ = execAgentRun(t, req, 15*time.Second, "kill", req.SessionIDA)
		_ = execAgentRun(t, req, 15*time.Second, "kill", req.SessionIDB)
		if attempt == maxAttempts {
			break
		}
	}
	resp.OpenA = openA
	resp.OpenB = openB

	resp.MetaA, resp.RunnerSessionIDA = readMeta(req, req.SessionIDA)
	resp.MetaB, resp.RunnerSessionIDB = readMeta(req, req.SessionIDB)
	resp.CodexIDs = listCodexIDs(req.CodexHome)
	for _, id := range resp.CodexIDs {
		path := findRolloutPath(req.CodexHome, id)
		resp.PromptByCodexID[id] = firstRealUserPrompt(path)
	}
	// Best-effort kill both sessions so the suite does not leak serves.
	_ = execAgentRun(t, req, 20*time.Second, "kill", req.SessionIDA)
	_ = execAgentRun(t, req, 20*time.Second, "kill", req.SessionIDB)
	resp.Elapsed = time.Since(start)
	return resp, nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not found in PATH")
	}
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
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
	if err := os.WriteFile(filepath.Join(req.Workspace, "README.md"), []byte("codex-concurrent-open-bind\n"), 0644); err != nil {
		return err
	}

	req.AgentRun, req.LLMMock, req.LLMMockRunCodex = ensureSessionBinaries(t, d, req.RepoRoot)
	if filepath.Dir(req.LLMMock) != filepath.Dir(req.LLMMockRunCodex) {
		return fmt.Errorf("llm-mock and llm-mock-run-codex must share a directory")
	}

	req.MockConfigFile = filepath.Join(req.TempDir, "llm-mock-config.json")
	writeConcurrentMockConfig(t, req.MockConfigFile)

	if req.SessionIDA == "" {
		req.SessionIDA = defaultSessionIDA
	}
	if req.SessionIDB == "" {
		req.SessionIDB = defaultSessionIDB
	}
	if req.PromptA == "" {
		req.PromptA = defaultPromptA
	}
	if req.PromptB == "" {
		req.PromptB = defaultPromptB
	}
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = defaultExecTimeout
	}
	return nil
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
