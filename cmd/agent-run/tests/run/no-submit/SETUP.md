# Scenario

**Feature**: `agent-run run --no-submit` (with required `--open`) stages a draft
in the TTY without auto-submitting a model turn

```
# validation
agent-run run --no-submit --agent-runner grok-tty "x" -> error (--no-submit requires --open)
agent-run run --open --no-submit --agent-runner fake-codex "x" -> error (non-TTY)

# open + no-submit (Option C: real grok under llm-mock-run-grok, no hooks)
agent-run run --agent-runner grok-tty --open --no-submit "draft-…"
  --agent-runner-binary llm-mock-run-grok
  --agent-runner-config-home <isolated>
  -> silent open start; draft NOT on argv; inject suffixCR=false
  -> auto-attach (AGENT_RUN_OPEN_ATTACH_INSTANT=1 in tests)
  -> on attach exit: stderr once "grok-tty: <id>"
  -> soft unbound OK if no provider session yet
  -> after settle: no mock HTTP chat / no session user_message for draft
```

## Preconditions

- Repository contains `cmd/agent-run`, `cmd/fake-codex`,
  `agent/llm/llm-mock` (server), and `agent/llm/llm-mock/llm-mock-run-grok`.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Non-TTY reject leaves use `fake-codex` on `PATH`.
- Option C leaves require real `grok` on `PATH`, build `llm-mock-run-grok` **and
  sibling `llm-mock`** (orchestrator resolves mock server as same-dir `llm-mock`),
  set isolated `--agent-runner-config-home`, and **strip**
  `AGENT_RUN_GROK_TTY_COMMAND` + `LLM_MOCK_RUN_GROK_COMMAND`.
- Option C leaves set `LLM_MOCK_RUN_FLAGS=--log-http <tmp>.jsonl` and
  `AGENT_RUN_OPEN_ATTACH_INSTANT=1`.
- Session-scoped build cache may share compiled binaries across parallel leaves:
  `$TMPDIR/agent-run-run-no-submit-doctest-<DOCTEST_SESSION_ID>/`
  (`agent-run`, `fake-codex`, `llm-mock`, `llm-mock-run-grok`).

## Steps

1. Root `Setup` resolves repo root, creates temp home + workspace, builds
   `agent-run`, `fake-codex`, `llm-mock`, and `llm-mock-run-grok` (session
   cache; `llm-mock` sibling required by the orchestrator), sets
   `AGENT_RUN_HOME` + `PATH`.
2. Grouping `Setup` narrows outcome class (`help` / `reject` / `with-open`) and
   runner / flags.
3. Leaf `Setup` finalizes flags, prompt, real-grok hooks strip, log-http path,
   and open-attach instant env.
4. `Run` executes `agent-run` with `req.Args`; for `open-real-grok-after`,
   settles then reads log-http + scans GROK_HOME for turn evidence.
5. Leaf `Assert` checks exit code, error text, help text, or no-turn / control
   turn oracles.

## Context

- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `DOCTEST_ROOT/../../../../..` (same depth as `run/open`).
- Session cache dir: `$TMPDIR/agent-run-run-no-submit-doctest-<DOCTEST_SESSION_ID>/`.
- Test hook env: `AGENT_RUN_OPEN_ATTACH_INSTANT=1` — auto-attach returns
  immediately so open leaves complete without interactive stdin.
- Default Option C draft marker: unique string unlikely to collide with UI chrome.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	grokTTYBannerMarker  = "GROK_TTY_BANNER"
	envOpenAttachInstant = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand    = "AGENT_RUN_GROK_TTY_COMMAND"
	envCodexTTYCommand   = "AGENT_RUN_CODEX_TTY_COMMAND"
	envLLMMockRunGrokCmd = "LLM_MOCK_RUN_GROK_COMMAND"
	envLLMMockRunFlags   = "LLM_MOCK_RUN_FLAGS"

	// defaultOptionCPrompt is a distinctive draft so HTTP/session scans stay precise.
	defaultOptionCPrompt = "draft-no-submit-OPTIONC-probe-zz9"
	defaultSettleAfter   = 5 * time.Second
	defaultRealGrokTimeout = 120 * time.Second
)

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "agent-run-run-no-submit-doctest-"+DOCTEST_SESSION_ID)
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

// ensureStubDist makes sure distDir has at least one embeddable file so
// //go:embed dist compiles. frontend-agent-run/dist is gitignored and may be
// absent in a fresh checkout; the CLI paths under test do not serve UI assets.
func ensureStubDist(distDir string) error {
	entries, statErr := os.ReadDir(distDir)
	if statErr == nil {
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			return nil
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(distDir, "index.html"), []byte("stub\n"), 0644)
}

func buildOnce(t *testing.T) (agentRun, fakeCodex, llmMockServer, llmMockRunGrok string, err error) {
	t.Helper()
	cache := sessionCacheDir()
	agentRun = filepath.Join(cache, "agent-run")
	fakeCodex = filepath.Join(cache, "fake-codex")
	// llm-mock must sit next to llm-mock-run-grok (mockServerExecutable sibling lookup).
	llmMockServer = filepath.Join(cache, "llm-mock")
	llmMockRunGrok = filepath.Join(cache, "llm-mock-run-grok")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot := filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(fakeCodex) &&
			fileExists(llmMockServer) && fileExists(llmMockRunGrok) {
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
		builds := []struct {
			out  string
			args []string
		}{
			{agentRun, []string{"build", "-o", agentRun, "./cmd/agent-run"}},
			{fakeCodex, []string{"build", "-o", fakeCodex, "./cmd/fake-codex"}},
			{llmMockServer, []string{"build", "-o", llmMockServer, "./agent/llm/llm-mock"}},
			{llmMockRunGrok, []string{"build", "-o", llmMockRunGrok, "./agent/llm/llm-mock/llm-mock-run-grok"}},
		}
		for _, b := range builds {
			cmd := exec.Command("go", b.args...)
			cmd.Dir = repoRoot
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("go %v: %w\n%s", b.args, err, string(out))
			}
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	return agentRun, fakeCodex, llmMockServer, llmMockRunGrok, err
}

func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

func fakeTUIHoldSeconds(sec int) string {
	return fmt.Sprintf(`sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep %d'`, sec)
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

func setEnvKV(req *Request, key, value string) {
	req.Env = withoutEnvKey(req.Env, key)
	req.Env = append(req.Env, key+"="+value)
}

func stripEnvKeys(req *Request, keys ...string) {
	for _, key := range keys {
		req.Env = withoutEnvKey(req.Env, key)
		// Explicit empty override so parent process env cannot reintroduce hooks
		// (exec Env last-wins for duplicates on Unix via Go os/exec).
		req.Env = append(req.Env, key+"=")
	}
}

func setGrokTTYCommand(req *Request, cmd string) {
	req.GrokTTYCommand = cmd
	setEnvKV(req, envGrokTTYCommand, cmd)
}

func setCodexTTYCommand(req *Request, cmd string) {
	req.CodexTTYCommand = cmd
	setEnvKV(req, envCodexTTYCommand, cmd)
}

func applyOpenInstantAttach(req *Request) {
	if !req.OpenInstantAttach {
		return
	}
	setEnvKV(req, envOpenAttachInstant, "1")
}

// stripOptionCHooks removes fake-TUI / mock-command overrides so real grok runs
// under llm-mock-run-grok (Option C contract).
func stripOptionCHooks(req *Request) {
	req.GrokTTYCommand = ""
	req.CodexTTYCommand = ""
	stripEnvKeys(req, envGrokTTYCommand, envCodexTTYCommand, envLLMMockRunGrokCmd)
}

func cleanupRegistryServes(home string) {
	for _, sub := range []string{"grok-tty-registry", "codex-tty-registry"} {
		dir := filepath.Join(home, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var ent struct {
				PID int `json:"pid"`
			}
			if json.Unmarshal(b, &ent) != nil || ent.PID <= 0 {
				continue
			}
			if ent.PID == os.Getpid() {
				continue
			}
			_ = exec.Command("kill", fmt.Sprint(ent.PID)).Run()
		}
	}
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	// Start from a cleaned process env so ambient hooks cannot reappear.
	base := os.Environ()
	for _, key := range []string{envGrokTTYCommand, envCodexTTYCommand, envLLMMockRunGrokCmd} {
		base = withoutEnvKey(base, key)
	}
	cmd.Env = append(base, env...)
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
	applyOpenInstantAttach(req)
	workDir := req.TempDir
	if req.Workspace != "" {
		workDir = req.Workspace
	}
	resp, err := execCmd(t, req.AgentRun, args, workDir, req.Env, req.ExecTimeout)
	if err != nil {
		return resp, err
	}
	if resp != nil {
		if id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty"); ok {
			resp.SessionID = id
		} else if id, ok := parsePrefixedSessionID(resp.Stderr, "codex-tty"); ok {
			resp.SessionID = id
		}
		needRegistry := req.Mode == "open-registry-after" || req.Mode == "open-snapshot-after" || req.Mode == "open-real-grok-after"
		if needRegistry && resp.SessionID != "" && resp.ExitCode == 0 {
			runner := req.Runner
			if runner == "" {
				runner = "grok-tty"
			}
			entry, rerr := readRegistryEntryOptional(req.Home, runner, resp.SessionID)
			if rerr == nil {
				resp.RegistryEntry = entry
			}
		}
		if req.Mode == "open-snapshot-after" && resp.RegistryEntry != nil && resp.RegistryEntry.ListenAddr != "" && resp.SessionID != "" {
			time.Sleep(400 * time.Millisecond)
			if text, serr := ttywatch.SnapshotText(resp.RegistryEntry.ListenAddr, resp.SessionID); serr == nil {
				resp.Snapshot = text
			}
		}
		if req.Mode == "open-real-grok-after" {
			settle := req.SettleAfter
			if settle <= 0 {
				settle = defaultSettleAfter
			}
			// Keep-alive PTY may still be finishing a buggy auto-submit turn.
			time.Sleep(settle)
			fillTurnOracles(t, req, resp)
		}
	}
	return resp, nil
}

func fillTurnOracles(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	if req.LogHTTPPath != "" {
		data, rerr := os.ReadFile(req.LogHTTPPath)
		if rerr == nil {
			resp.LogHTTPContent = string(data)
			for _, line := range strings.Split(resp.LogHTTPContent, "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					resp.LogHTTPLines = append(resp.LogHTTPLines, line)
				}
			}
		}
		resp.HasMockChatHTTP = logHTTPIndicatesChatTurn(resp.LogHTTPContent, req.Prompt)
	}
	if req.GrokHome != "" {
		found, n := scanGrokHomeUserMessage(req.GrokHome, req.Prompt)
		resp.HasUserMessageForPrompt = found
		resp.GrokSessionsScanned = n
	}
}

// logHTTPIndicatesChatTurn reports mock chat activity for the draft.
// Matches /v1/chat/completions (or responses) exchanges that mention the prompt
// when prompt is non-empty; any chat path counts when prompt is empty.
func logHTTPIndicatesChatTurn(content, prompt string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	prompt = strings.TrimSpace(prompt)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		isChat := strings.Contains(lower, "/v1/chat/completions") ||
			strings.Contains(lower, "/chat/completions") ||
			strings.Contains(lower, "/v1/responses")
		if !isChat {
			// Fallback: any JSONL exchange whose body embeds the draft.
			if prompt != "" && strings.Contains(line, prompt) {
				return true
			}
			continue
		}
		if prompt == "" || strings.Contains(line, prompt) {
			return true
		}
	}
	return false
}

// scanGrokHomeUserMessage walks GROK_HOME/sessions for updates.jsonl lines with
// user_message_chunk whose text contains wantPrompt.
func scanGrokHomeUserMessage(grokHome, wantPrompt string) (found bool, sessions int) {
	wantPrompt = strings.TrimSpace(wantPrompt)
	root := filepath.Join(grokHome, "sessions")
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			// Count leaf-ish session dirs that contain updates.jsonl later.
			return nil
		}
		base := d.Name()
		if base != "updates.jsonl" && base != "events.jsonl" {
			return nil
		}
		if base == "updates.jsonl" {
			sessions++
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		text := string(data)
		if wantPrompt != "" && !strings.Contains(text, wantPrompt) {
			// Still accept user_message_chunk without exact prompt when scanning
			// only for "any" turn — callers pass distinctive prompts.
			return nil
		}
		if base == "updates.jsonl" {
			for _, line := range strings.Split(text, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.Contains(line, "user_message_chunk") &&
					(wantPrompt == "" || strings.Contains(line, wantPrompt)) {
					found = true
					return filepath.SkipAll
				}
			}
		}
		// events.jsonl: user role / turn_started with prompt
		if base == "events.jsonl" && wantPrompt != "" && strings.Contains(text, wantPrompt) {
			if strings.Contains(text, "user") || strings.Contains(text, "turn_started") {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found, sessions
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

// parsePrefixedSessionID extracts "<runner>: <id>" where id is a single token.
// Skips multi-word diagnostic lines (e.g. "grok-tty: grok session …").
func parsePrefixedSessionID(stderr, runner string) (string, bool) {
	prefix := runner + ":"
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if rest == "" || strings.Contains(rest, " ") {
			continue
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
			return rest, true
		}
	}
	return "", false
}

func countPrefixedSessionIDLines(stderr, runner string) int {
	prefix := runner + ":"
	n := 0
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if rest == "" || strings.Contains(rest, " ") {
			continue
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
			n++
		}
	}
	return n
}

func registryDir(home, runner string) string {
	return filepath.Join(home, runner+"-registry")
}

func registryPath(home, runner, sessionID string) string {
	return filepath.Join(registryDir(home, runner), sessionID+".json")
}

func readRegistryEntryOptional(home, runner, sessionID string) (*RegistryEntry, error) {
	path := registryPath(home, runner, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entry RegistryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if entry.SessionID == "" {
		entry.SessionID = sessionID
	}
	return &entry, nil
}

func portOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// configureRealGrokOpen wires Option C defaults: llm-mock-run-grok binary,
// isolated config home, log-http, no hooks, instant attach.
func configureRealGrokOpen(t *testing.T, req *Request) error {
	t.Helper()
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skipf("grok not found in PATH: %v", err)
	}
	if strings.TrimSpace(req.LLMMockRunGrok) == "" || !fileExists(req.LLMMockRunGrok) {
		return fmt.Errorf("LLMMockRunGrok missing; root Setup must build llm-mock-run-grok")
	}
	if req.Workspace == "" {
		req.Workspace = filepath.Join(req.TempDir, "workspace")
		if err := os.MkdirAll(req.Workspace, 0755); err != nil {
			return fmt.Errorf("mkdir workspace: %w", err)
		}
	}
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, ".grok")
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return fmt.Errorf("mkdir grok home: %w", err)
	}
	if req.LogHTTPPath == "" {
		req.LogHTTPPath = filepath.Join(req.TempDir, "option-c-http.jsonl")
	}
	// Ensure parent exists; mock appends the file on first request.
	if err := os.MkdirAll(filepath.Dir(req.LogHTTPPath), 0755); err != nil {
		return err
	}
	// Truncate prior content if reusing path.
	_ = os.WriteFile(req.LogHTTPPath, nil, 0644)

	req.Runner = "grok-tty"
	req.AgentRunnerBinary = req.LLMMockRunGrok
	req.OpenInstantAttach = true
	req.Mode = "open-real-grok-after"
	if req.SettleAfter <= 0 {
		req.SettleAfter = defaultSettleAfter
	}
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = defaultRealGrokTimeout
	}
	if strings.TrimSpace(req.Prompt) == "" {
		req.Prompt = defaultOptionCPrompt
	}

	stripOptionCHooks(req)
	setEnvKV(req, envOpenAttachInstant, "1")
	// LLM_MOCK_RUN_FLAGS prepended by llm-mock-run-grok; path must be space-free.
	setEnvKV(req, envLLMMockRunFlags, "--log-http "+req.LogHTTPPath)

	t.Cleanup(func() { cleanupRegistryServes(req.Home) })
	return nil
}

func openRealGrokArgs(req *Request) []string {
	args := []string{
		"run",
		"--agent-runner", "grok-tty",
		"--open",
		"--dir", req.Workspace,
		"--agent-runner-binary", req.AgentRunnerBinary,
		"--agent-runner-config-home", req.GrokHome,
	}
	if req.NoSubmit {
		args = append(args, "--no-submit")
	}
	if strings.TrimSpace(req.Prompt) != "" {
		args = append(args, req.Prompt)
	}
	return args
}

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.Workspace = filepath.Join(req.TempDir, "workspace")
	for _, d := range []string{req.Home, req.Workspace} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}
	agentRun, fakeCodex, llmMockServer, llmMockRunGrok, err := buildOnce(t)
	if err != nil {
		return err
	}
	binDir := filepath.Join(req.TempDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	req.AgentRun = filepath.Join(binDir, "agent-run")
	req.FakeCodex = filepath.Join(binDir, "fake-codex")
	// Sibling pair: llm-mock-run-grok resolves mock server as same-dir llm-mock.
	req.LLMMockRunGrok = filepath.Join(binDir, "llm-mock-run-grok")
	llmMockDst := filepath.Join(binDir, "llm-mock")
	for _, pair := range []struct{ src, dst string }{
		{agentRun, req.AgentRun},
		{fakeCodex, req.FakeCodex},
		{llmMockServer, llmMockDst},
		{llmMockRunGrok, req.LLMMockRunGrok},
	} {
		if out, err := exec.Command("cp", pair.src, pair.dst).CombinedOutput(); err != nil {
			return fmt.Errorf("cp %s: %w\n%s", filepath.Base(pair.dst), err, string(out))
		}
		if err := os.Chmod(pair.dst, 0755); err != nil {
			return err
		}
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+binDir+":"+os.Getenv("PATH"),
	)
	// Default strip ambient hooks at root so help/reject stay clean too.
	stripEnvKeys(req, envGrokTTYCommand, envLLMMockRunGrokCmd)
	req.Args = []string{"run"}
	return nil
}
```
