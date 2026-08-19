# Scenario

**Feature**: `codex-tty` runner — adhoc ptywrap server, registry, scrollback capture, attach

```
agent-run run --agent-runner codex-tty "prompt"
  -> adhoc ptywrap on 127.0.0.1:0
  -> registry codex-tty-registry/<id>.json
  -> stderr codex-tty: session-N
  -> capture sidecar (readonly scrollback)
  -> inject prompt after CODEX_TTY_BANNER
  -> events under sessions/codex-tty/

agent-run attach <id> -> registry lookup -> ptyclient WS attach
```

## Preconditions

- Repository contains `cmd/agent-run` (build may fail until codex-tty is implemented).
- Each test uses isolated `AGENT_RUN_HOME=filepath.Join(t.TempDir(), ".agent-run")`.
- Default-suite tests set `AGENT_RUN_CODEX_TTY_COMMAND` to a fake interactive TUI script
  (not a non-interactive Codex JSON mode).
- No `agent-term serve` — each run owns its own ephemeral ptywrap HTTP server.
- Fake TUI scripts print `CODEX_TTY_BANNER` before the prompt marker so banner-wait
  logic is deterministic.

## Steps

1. Root `Setup` builds `agent-run`, sets `AGENT_RUN_HOME`, configures fake TUI env
   unless `req.SkipFakeTUI` is set by a real Codex scenario.
2. Grouping `Setup` narrows subcommand (`run`, `attach`, `help`) or real-codex profile.
3. Leaf `Setup` sets prompt, fake TUI variant, or starts a background codex-tty run.
4. `Run` executes CLI or performs registry/attach probe against a live session.
5. Leaf `Assert` checks stderr session id, registry JSON, captured output, or attach.

```go
import (
	"runtime"
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
	"regexp"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"

	"github.com/creack/pty"
	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
	"github.com/xhd2015/doctest/session"
)

const codexTTYBannerMarker = "CODEX_TTY_BANNER"

func fakeTUIRespondHi() string {
	// Codex injects the prompt (no argv). Hold before read so inject wins;
	// sleep after echo so snapshotHold can catch Response before serve teardown.
	return `sh -c 'printf "CODEX_TTY_BANNER\nCodex › "; sleep 0.3; read line; echo "Response: $line"; sleep 0.25'`
}

func fakeTUILongRun() string {
	return `sh -c 'printf "CODEX_TTY_BANNER\n"; sleep 30'`
}

func fakeTUICodexResumeThenSleep(seconds int) string {
	return fmt.Sprintf(`sh -c 'printf "CODEX_TTY_BANNER\nTo continue this session, run codex resume %%s\n" "$FAKE_CODEX_SESSION_ID"; sleep %d'`, seconds)
}

func fakeTUICodexResumeWithFallback(fallback string) string {
	return `sh -c 'printf "CODEX_TTY_BANNER\nTo continue this session, run codex resume %s\nCodex › " "$FAKE_CODEX_SESSION_ID"; read line; printf "%s\n" "$FAKE_CODEX_FALLBACK_TEXT"'`
}

func fakeTUIDelayedBanner() string {
	return `sh -c 'sleep 0.3; printf "CODEX_TTY_BANNER\nCodex › "; read line; echo "Response: $line"'`
}

func fakeTUIRequiresCR() string {
	return `sh -c 'printf "CODEX_TTY_BANNER\nCodex › "; read -r line; echo "SUBMITTED:$line"'`
}

func fakeTUIRawCodexScrollback() string {
	return `sh -c 'printf "CODEX_TTY_BANNER\n"; read -r line; printf ">4;0m>7u\n╭────────────────────────╮\n│ >_ OpenAI Codex        │\n│ model: loading /model to change │\n│ directory: /tmp/work   │\n╰────────────────────────╯\nStarting MCP servers...\nBooting MCP...\nWorking...\nWorking...\n› %s\nls output:\nAGENTS.md\ncmd\npkgs\n" "$line"'`
}

func codexTTYRegistryDir(home string) string {
	return filepath.Join(home, "codex-tty-registry")
}

func codexTTYRegistryPath(home, sessionID string) string {
	return filepath.Join(codexTTYRegistryDir(home), sessionID+".json")
}

func grokTTYRegistryPath(home, sessionID string) string {
	return filepath.Join(home, "grok-tty-registry", sessionID+".json")
}

func reserveUnusedLocalAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve unused localhost address: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close reserved listener: %v", err)
	}
	return addr
}

func writeStaleGrokTTYRegistry(t *testing.T, home, sessionID string) string {
	t.Helper()
	path := grokTTYRegistryPath(home, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir stale grok registry dir: %v", err)
	}
	addr := reserveUnusedLocalAddr(t)
	entry := CodexTTYRegistryEntry{
		SessionID:  sessionID,
		ListenAddr: addr,
		PID:        -1,
		CreatedAt:  time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal stale grok registry: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write stale grok registry: %v", err)
	}
	return addr
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 45 * time.Second
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
	appendCodexTTYEnv(req)
	appendCodexHomeEnv(req)
	return execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
}

func appendCodexHomeEnv(req *Request) {
	if strings.TrimSpace(req.CodexHome) == "" {
		return
	}
	req.Env = withoutEnvKey(req.Env, "CODEX_HOME")
	req.Env = append(req.Env, "CODEX_HOME="+req.CodexHome)
	req.Env = withoutEnvKey(req.Env, "HOME")
	req.Env = append(req.Env, "HOME="+filepath.Dir(req.CodexHome))
}

func appendCodexTTYEnv(req *Request) {
	if req.SkipFakeTUI {
		return
	}
	cmd := strings.TrimSpace(req.CodexTTYCommand)
	if cmd == "" {
		cmd = fakeTUIRespondHi()
	}
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_CODEX_TTY_COMMAND")
	req.Env = append(req.Env, "AGENT_RUN_CODEX_TTY_COMMAND="+cmd)
}

func setCodexTTYCommand(req *Request, cmd string) {
	req.CodexTTYCommand = cmd
	appendCodexTTYEnv(req)
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

func parseCodexTTYSessionID(stderr string) (string, bool) {
	re := regexp.MustCompile(`codex-tty:\s*(session-\d+)`)
	m := re.FindStringSubmatch(stderr)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func waitForCodexTTYSessionLine(t *testing.T, buf *bytes.Buffer, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if id, ok := parseCodexTTYSessionID(buf.String()); ok {
			return id
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for codex-tty session id in stderr:\n%s", buf.String())
	return ""
}

func waitForRegistryEntry(t *testing.T, home, sessionID string, timeout time.Duration) *CodexTTYRegistryEntry {
	t.Helper()
	deadline := time.Now().Add(timeout)
	path := codexTTYRegistryPath(home, sessionID)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil {
			var entry CodexTTYRegistryEntry
			if json.Unmarshal(data, &entry) == nil && entry.ListenAddr != "" {
				return &entry
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for registry file %s", path)
	return nil
}

func portOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func waitForTCPAddr(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if portOpen(addr) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for TCP %s", addr)
}

func probeAttachWS(t *testing.T, listenAddr, sessionID string) (bool, string) {
	t.Helper()
	base := "http://" + listenAddr
	c := ptyclient.NewClient(base)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := ptyclient.Attach(c, ptyclient.ConnectOptions{
			SessionID:      sessionID,
			AttachSnapshot: true,
			Wait:           false,
		})
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			return false, err.Error()
		}
		return true, ""
	case <-ctx.Done():
		return true, ""
	}
}

func probeAttachCLI(t *testing.T, req *Request, sessionID string) (*Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, "attach", sessionID)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start attach under pty: %w", err)
	}
	defer func() { _ = ptmx.Close() }()

	var output bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, ptmx)
		close(readDone)
	}()

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	var waitErr error
	select {
	case waitErr = <-waitDone:
	case <-ctx.Done():
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		waitErr = ctx.Err()
	}
	_ = ptmx.Close()
	select {
	case <-readDone:
	case <-time.After(500 * time.Millisecond):
	}
	resp := &Response{
		Stdout: output.String(),
		Stderr: output.String(),
		Err:    waitErr,
	}
	if waitErr == nil {
		resp.AttachProbeOK = true
		return resp, nil
	}
	if ctx.Err() != nil {
		if strings.Contains(output.String(), "codex-tty") || output.Len() > 0 || ctx.Err() == context.DeadlineExceeded {
			resp.AttachProbeOK = true
			return resp, nil
		}
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, waitErr
}

func startCodexTTYBackground(t *testing.T, req *Request) {
	t.Helper()
	appendCodexTTYEnv(req)
	args := []string{"run", "--agent-runner", "codex-tty"}
	if req.CodexTTYPrompt != "" {
		args = append(args, req.CodexTTYPrompt)
	} else {
		args = append(args, "hold")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	req.BackgroundStderr = &bytes.Buffer{}
	req.BackgroundStdout = &bytes.Buffer{}
	cmd.Stderr = req.BackgroundStderr
	cmd.Stdout = req.BackgroundStdout
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start background codex-tty run: %v", err)
	}
	req.BackgroundCmd = cmd
	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
	req.CodexTTYSessionID = waitForCodexTTYSessionLine(t, req.BackgroundStderr, 15*time.Second)
}

func runRegistryWhileRunning(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CodexTTYSessionID == "" {
		t.Fatal("registry-while-running requires CodexTTYSessionID from background Setup")
	}
	entry := waitForRegistryEntry(t, req.Home, req.CodexTTYSessionID, 10*time.Second)
	waitForTCPAddr(t, entry.ListenAddr, 5*time.Second)
	return &Response{RegistryEntry: entry}, nil
}

func runAttachProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CodexTTYSessionID == "" {
		t.Fatal("attach-probe requires CodexTTYSessionID from background Setup")
	}
	entry := waitForRegistryEntry(t, req.Home, req.CodexTTYSessionID, 10*time.Second)
	ok, probeErr := probeAttachWS(t, entry.ListenAddr, entry.SessionID)
	return &Response{
		RegistryEntry:  entry,
		AttachProbeOK:  ok,
		AttachProbeErr: probeErr,
	}, nil
}

func runAttachCLIOnlyProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CodexTTYSessionID == "" {
		t.Fatal("attach-cli-only-probe requires CodexTTYSessionID from background Setup")
	}
	resp, err := probeAttachCLI(t, req, req.CodexTTYSessionID)
	if resp == nil {
		resp = &Response{}
	}
	resp.BackgroundStdout = req.BackgroundStdout.String()
	resp.BackgroundStderr = req.BackgroundStderr.String()
	if entry, readErr := waitForRegistryEntryNoFatal(req.Home, req.CodexTTYSessionID, 2*time.Second); readErr == nil {
		resp.RegistryEntry = entry
	}
	return resp, err
}

func waitForRegistryEntryNoFatal(home, sessionID string, timeout time.Duration) (*CodexTTYRegistryEntry, error) {
	deadline := time.Now().Add(timeout)
	path := codexTTYRegistryPath(home, sessionID)
	var lastErr error
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			var entry CodexTTYRegistryEntry
			if err := json.Unmarshal(data, &entry); err == nil && entry.ListenAddr != "" {
				return &entry, nil
			} else if err != nil {
				lastErr = err
			}
		} else {
			lastErr = err
		}
		time.Sleep(50 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("registry entry %s not found", path)
	}
	return nil, lastErr
}

func collectAttachScrollback(t *testing.T, listenAddr, sessionID string, collectFor time.Duration) (string, error) {
	t.Helper()
	base := "http://" + listenAddr
	c := ptyclient.NewClient(base)
	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), collectFor+5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := ptyclient.AttachWithIO(c, ptyclient.ConnectOptions{
			SessionID:      sessionID,
			AttachSnapshot: true,
			Wait:           true,
			SkipTTYCheck:   true,
		}, strings.NewReader(""), &buf, io.Discard)
		done <- err
	}()
	select {
	case err := <-done:
		return buf.String(), err
	case <-time.After(collectFor):
		return buf.String(), nil
	case <-ctx.Done():
		return buf.String(), ctx.Err()
	}
}

func runAttachScrollbackProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CodexTTYSessionID == "" {
		t.Fatal("attach-scrollback-probe requires CodexTTYSessionID from background Setup")
	}
	entry := waitForRegistryEntry(t, req.Home, req.CodexTTYSessionID, 10*time.Second)
	scrollback, err := collectAttachScrollback(t, entry.ListenAddr, entry.SessionID, 2*time.Second)
	resp := &Response{
		RegistryEntry:    entry,
		AttachScrollback: scrollback,
		AttachProbeOK:    entry.ListenAddr != "",
	}
	if err != nil && !strings.Contains(scrollback, codexTTYBannerMarker) {
		resp.AttachProbeErr = err.Error()
		resp.AttachProbeOK = false
	}
	return resp, nil
}

func runAttachInteractiveProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CodexTTYSessionID == "" {
		t.Fatal("attach-interactive-probe requires CodexTTYSessionID from background Setup")
	}
	resp, err := probeAttachCLI(t, req, req.CodexTTYSessionID)
	if resp != nil && resp.AttachProbeOK {
		entry := waitForRegistryEntry(t, req.Home, req.CodexTTYSessionID, 10*time.Second)
		resp.RegistryEntry = entry
		resp.BackgroundStdout = req.BackgroundStdout.String()
		resp.BackgroundStderr = req.BackgroundStderr.String()
		return resp, err
	}
	entry := waitForRegistryEntry(t, req.Home, req.CodexTTYSessionID, 10*time.Second)
	ok, probeErr := probeAttachWS(t, entry.ListenAddr, entry.SessionID)
	if resp == nil {
		resp = &Response{}
	}
	resp.AttachProbeOK = ok
	resp.AttachProbeErr = probeErr
	resp.RegistryEntry = entry
	resp.BackgroundStdout = req.BackgroundStdout.String()
	resp.BackgroundStderr = req.BackgroundStderr.String()
	return resp, err
}

func findCodexTTYEventsJSONL(t *testing.T, home string) (string, []string) {
	t.Helper()
	// Flat AGENT_RUN_HOME/sessions/<sessionID>/ layout.
	candidates := []string{
		filepath.Join(home, "sessions"),
	}
	var found string
	var lines []string
	for _, root := range candidates {
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if info.Name() == "events.jsonl" {
				found = path
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					t.Fatalf("read %s: %v", path, readErr)
				}
				lines = nil
				for _, line := range strings.Split(string(data), "\n") {
					if strings.TrimSpace(line) != "" {
						lines = append(lines, line)
					}
				}
				return filepath.SkipAll
			}
			return nil
		})
		if found != "" {
			break
		}
	}
	return found, lines
}

func eventsContainSubstring(t *testing.T, lines []string, want string) bool {
	t.Helper()
	for _, line := range lines {
		var ev map[string]any
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		blob, _ := json.Marshal(ev)
		if strings.Contains(strings.ToLower(string(blob)), strings.ToLower(want)) {
			return true
		}
	}
	return false
}

func codexRolloutPath(codexHome, sessionID string) string {
	return filepath.Join(codexHome, "sessions", "2026", "07", "02", "rollout-2026-07-02T12-01-53-"+sessionID+".jsonl")
}

func ensureCodexTranscriptPath(t *testing.T, req *Request) string {
	t.Helper()
	if req.CodexHome == "" {
		req.CodexHome = filepath.Join(req.TempDir, ".codex")
	}
	if req.CodexTranscriptSessionID == "" {
		req.CodexTranscriptSessionID = "019f20fd-8569-7910-ab0b-9d898d66e3e6"
	}
	if req.CodexTranscriptPath == "" {
		req.CodexTranscriptPath = codexRolloutPath(req.CodexHome, req.CodexTranscriptSessionID)
	}
	if err := os.MkdirAll(filepath.Dir(req.CodexTranscriptPath), 0755); err != nil {
		t.Fatalf("mkdir codex transcript dir: %v", err)
	}
	return req.CodexTranscriptPath
}

func appendCodexTranscriptJSONL(path string, lines ...string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

func codexSessionMetaLine(sessionID, cwd string) string {
	line, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-02T12:01:53Z",
		"type":      "session_meta",
		"payload":   map[string]any{"session_id": sessionID, "cwd": cwd},
	})
	return string(line)
}

func codexAgentMessageLine(text string) string {
	line, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-02T12:01:54Z",
		"type":      "event_msg",
		"payload":   map[string]any{"type": "agent_message", "message": text},
	})
	return string(line)
}

func codexTaskCompleteLine(text string) string {
	line, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-02T12:01:55Z",
		"type":      "event_msg",
		"payload":   map[string]any{"type": "task_complete", "last_agent_message": text},
	})
	return string(line)
}

func codexAssistantResponseItemLine(texts ...string) string {
	content := make([]map[string]any, 0, len(texts))
	for _, text := range texts {
		content = append(content, map[string]any{"type": "output_text", "text": text})
	}
	line, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-02T12:01:54Z",
		"type":      "response_item",
		"payload": map[string]any{
			"type":    "message",
			"role":    "assistant",
			"content": content,
		},
	})
	return string(line)
}

func codexFunctionCallOutputLine(output string) string {
	line, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-02T12:01:54Z",
		"type":      "response_item",
		"payload":   map[string]any{"type": "function_call_output", "output": output},
	})
	return string(line)
}

func seedCodexTranscript(t *testing.T, req *Request, lines ...string) {
	t.Helper()
	path := ensureCodexTranscriptPath(t, req)
	seed := []string{codexSessionMetaLine(req.CodexTranscriptSessionID, req.TempDir)}
	seed = append(seed, lines...)
	if err := appendCodexTranscriptJSONL(path, seed...); err != nil {
		t.Fatalf("seed codex transcript: %v", err)
	}
}

func scheduleCodexTranscriptAppends(t *testing.T, req *Request) {
	t.Helper()
	if len(req.CodexTranscriptSchedules) == 0 {
		return
	}
	path := ensureCodexTranscriptPath(t, req)
	for _, schedule := range req.CodexTranscriptSchedules {
		s := schedule
		go func() {
			time.Sleep(s.Delay)
			if err := appendCodexTranscriptJSONL(path, s.Lines...); err != nil {
				t.Errorf("append codex transcript: %v", err)
			}
		}()
	}
}

func runCodexJSONLStreamProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	appendCodexTTYEnv(req)
	appendCodexHomeEnv(req)
	scheduleCodexTranscriptAppends(t, req)

	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.AgentRun, req.Args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	want := req.StreamProbeSubstring
	probeTimeout := req.StreamProbeTimeout
	if probeTimeout <= 0 {
		probeTimeout = 8 * time.Second
	}
	resp := &Response{}
	deadline := time.After(probeTimeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	var waitErr error
	waited := false
	for !waited {
		select {
		case waitErr = <-waitDone:
			waited = true
		case <-ticker.C:
			if want != "" && strings.Contains(stdout.String(), want) {
				resp.StreamProbeSeen = true
				resp.StreamProbeBeforeExit = true
			}
		case <-deadline:
			if want != "" && strings.Contains(stdout.String(), want) {
				resp.StreamProbeSeen = true
			}
		case <-ctx.Done():
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			waitErr = ctx.Err()
			waited = true
		}
		if resp.StreamProbeBeforeExit && want != "" {
			break
		}
	}
	if !waited {
		waitErr = <-waitDone
	}

	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.StreamProbeStdout = stdout.String()
	resp.Err = waitErr
	if waitErr == nil {
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, waitErr
}

func runConcurrentCodexTTYRuns(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	appendCodexTTYEnv(req)
	appendCodexHomeEnv(req)
	count := req.ConcurrentRuns
	if count <= 0 {
		count = 3
	}
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type runResult struct {
		index  int
		stdout string
		stderr string
		err    error
	}
	results := make(chan runResult, count)
	start := make(chan struct{})
	for i := 0; i < count; i++ {
		i := i
		go func() {
			<-start
			cmd := exec.CommandContext(ctx, req.AgentRun, req.Args...)
			cmd.Dir = req.TempDir
			cmd.Env = append(os.Environ(), req.Env...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			results <- runResult{
				index:  i,
				stdout: stdout.String(),
				stderr: stderr.String(),
				err:    err,
			}
		}()
	}
	close(start)

	resp := &Response{
		ConcurrentSessionIDs: make([]string, count),
		ConcurrentStderrs:    make([]string, count),
		ConcurrentStdouts:    make([]string, count),
	}
	var firstErr error
	for i := 0; i < count; i++ {
		result := <-results
		resp.ConcurrentStdouts[result.index] = result.stdout
		resp.ConcurrentStderrs[result.index] = result.stderr
		if firstErr == nil && result.err != nil {
			firstErr = result.err
		}
		if id, ok := parseCodexTTYSessionID(result.stderr); ok {
			resp.ConcurrentSessionIDs[result.index] = id
		}
	}
	resp.Err = firstErr
	if firstErr != nil {
		var joined strings.Builder
		for i := range resp.ConcurrentStderrs {
			fmt.Fprintf(&joined, "[%d] stdout:\n%s\n[%d] stderr:\n%s\n", i, resp.ConcurrentStdouts[i], i, resp.ConcurrentStderrs[i])
		}
		resp.Stdout = joined.String()
		return resp, nil
	}
	return resp, nil
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("expected exit code %d, got %d, stderr:\n%s", want, resp.ExitCode, resp.Stderr)
	}
}

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
	appendCodexTTYEnv(req)
	return nil
}

func httpGetHealth(listenAddr string) (int, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + listenAddr + "/api/terminal/sessions")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

// ensureLLMMockCodex builds llm-mock-run-codex and sibling llm-mock into req.TempDir/bin.
// Sibling llm-mock is required: orchestrator looks next to the run-codex binary for the server.
func ensureLLMMockCodex(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.TempDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	req.LLMMockRunCodex = filepath.Join(binDir, "llm-mock-run-codex")
	req.LLMMockServer = filepath.Join(binDir, "llm-mock")
	builds := []struct {
		out  string
		pkg  string
	}{
		{req.LLMMockRunCodex, "./agent/llm/llm-mock/llm-mock-run-codex"},
		{req.LLMMockServer, "./agent/llm/llm-mock"},
	}
	for _, b := range builds {
		if _, err := os.Stat(b.out); err == nil {
			continue
		}
		cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", b.out, b.pkg)
		cmd.Dir = req.RepoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v\n%s", b.pkg, err, out)
		}
	}
}

func writeDefaultMockCodexConfig(t *testing.T, req *Request) {
	t.Helper()
	if req.MockConfigFile == "" {
		req.MockConfigFile = filepath.Join(req.TempDir, "llm-mock-config.json")
	}
	const body = `{
  "exchanges": [
    {
      "request": {"role": "user", "content": "follow-up-two", "index": -1},
      "response": {"content": "SECOND_MOCK_REPLY", "finish_reason": "stop"}
    },
    {
      "request": {"role": "user", "content": "draft-only-text", "index": -1},
      "response": {"content": "SHOULD_NOT_SEE_DRAFT_REPLY", "finish_reason": "stop"}
    }
  ]
}
`
	if err := os.WriteFile(req.MockConfigFile, []byte(body), 0644); err != nil {
		t.Fatalf("write mock config: %v", err)
	}
	req.Env = withoutEnvKey(req.Env, "LLM_MOCK_CONFIG_FILE")
	req.Env = withoutEnvKey(req.Env, "LLM_MOCK_CONFIG")
	req.Env = append(req.Env, "LLM_MOCK_CONFIG_FILE="+req.MockConfigFile)
}

func plainSnapshotText(raw string) string {
	// Strip common CSI / OSC for assertions.
	reCSI := regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	reOSC := regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)
	s := reCSI.ReplaceAllString(raw, "")
	s = reOSC.ReplaceAllString(s, "")
	reOther := regexp.MustCompile(`\x1b.`)
	return reOther.ReplaceAllString(s, "")
}

func execAgentRunEnv(t *testing.T, req *Request, timeout time.Duration, args ...string) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return execCmd(t, req.AgentRun, args, req.WorkspaceDir, req.Env, timeout)
}

func ttySnapshot(t *testing.T, req *Request, sessionID string) string {
	t.Helper()
	resp, err := execAgentRunEnv(t, req, 15*time.Second, "tty", "snapshot", sessionID)
	if err != nil {
		t.Logf("tty snapshot err: %v stdout=%s stderr=%s", err, resp.Stdout, resp.Stderr)
	}
	if resp == nil {
		return ""
	}
	return resp.Stdout
}

func ttyStatus(t *testing.T, req *Request, sessionID string) string {
	t.Helper()
	resp, err := execAgentRunEnv(t, req, 15*time.Second, "tty", "status", sessionID)
	if err != nil && resp == nil {
		return err.Error()
	}
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr
}

func runMockUIOpenIdle(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("mock-ui-%d", time.Now().UnixNano())
	}
	if req.WorkspaceDir == "" {
		req.WorkspaceDir = filepath.Join(req.TempDir, "ws")
		if err := os.MkdirAll(req.WorkspaceDir, 0755); err != nil {
			return nil, err
		}
		// git init helps some codex workspace checks; ignore errors.
		_ = exec.Command("git", "init", "-q", req.WorkspaceDir).Run()
		_ = os.WriteFile(filepath.Join(req.WorkspaceDir, "README.md"), []byte("ok\n"), 0644)
	}
	args := []string{
		"run",
		"--agent-runner", "codex-tty",
		"--agent-runner-binary", req.LLMMockRunCodex,
		"--session-id", req.SessionID,
		"--dir", req.WorkspaceDir,
		"--open",
	}
	openResp, err := execAgentRunEnv(t, req, 90*time.Second, args...)
	out := &Response{
		SessionID: req.SessionID,
		Stdout:    "",
		Stderr:    "",
	}
	if openResp != nil {
		out.Stdout = openResp.Stdout
		out.Stderr = openResp.Stderr
		out.ExitCode = openResp.ExitCode
	}
	if err != nil {
		out.Err = err
		return out, err
	}
	// Poll until trust is gone and sendable (or timeout).
	deadline := time.Now().Add(60 * time.Second)
	var snap, status string
	for time.Now().Before(deadline) {
		snap = ttySnapshot(t, req, req.SessionID)
		status = ttyStatus(t, req, req.SessionID)
		plain := plainSnapshotText(snap)
		low := strings.ToLower(plain)
		trustVisible := strings.Contains(low, "do you trust the contents") ||
			strings.Contains(low, "yes, continue")
		if !trustVisible && (strings.Contains(status, "sendable: yes") ||
			(strings.Contains(plain, "›") && strings.Contains(low, "openai codex"))) {
			break
		}
		time.Sleep(2 * time.Second)
	}
	out.Snapshot = snap
	out.StatusText = status
	return out, nil
}

func runMockUISend(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	openResp, err := runMockUIOpenIdle(t, req)
	if err != nil {
		return openResp, err
	}
	sendArgs := []string{"send", "--max-wait", "45s", req.SessionID}
	if req.MockUINoSubmit {
		sendArgs = []string{"send", "--no-submit", "--max-wait", "20s", req.SessionID}
	}
	text := req.MockUISendText
	if text == "" {
		text = "follow-up-two"
	}
	sendArgs = append(sendArgs, text)
	sendResp, sendErr := execAgentRunEnv(t, req, 60*time.Second, sendArgs...)
	if openResp == nil {
		openResp = &Response{SessionID: req.SessionID}
	}
	if sendResp != nil {
		openResp.SendStdout = sendResp.Stdout
		openResp.SendStderr = sendResp.Stderr
		openResp.ExitCode = sendResp.ExitCode
	}
	// Poll until submit evidence appears (composer cleared / assistant chrome), or timeout.
	deadline := time.Now().Add(45 * time.Second)
	if req.MockUINoSubmit {
		deadline = time.Now().Add(4 * time.Second)
	}
	var snap string
	for {
		snap = ttySnapshot(t, req, req.SessionID)
		if req.MockUINoSubmit {
			break
		}
		if mockUISubmitted(snap, text, req.MockUIExpectReply) {
			break
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}
	openResp.Snapshot = snap
	openResp.StatusText = ttyStatus(t, req, req.SessionID)
	if sendErr != nil {
		openResp.Err = sendErr
		return openResp, sendErr
	}
	return openResp, nil
}

// mockUISubmitted reports whether scrollback shows the send was submitted (not draft-only).
func mockUISubmitted(raw, sendText, expectReply string) bool {
	plain := plainSnapshotText(raw)
	low := strings.ToLower(plain)
	if expectReply != "" && strings.Contains(plain, expectReply) {
		return true
	}
	if strings.Contains(low, "no matching exchange") {
		return true
	}
	if strings.Contains(plain, "•") || strings.Contains(low, "esc to interrupt") {
		return true
	}
	// Composer empty after the user line: submitted, awaiting/idle.
	// Draft-only failure mashes text into the footer line: "› follow-up-twogpt-…"
	if sendText != "" && strings.Contains(plain, sendText) {
		if strings.Contains(plain, "›"+sendText) || strings.Contains(plain, "› "+sendText) {
			// still looks like active draft if glued to model footer without newline gap
			idx := strings.Index(plain, sendText)
			if idx >= 0 {
				after := plain[idx+len(sendText):]
				if strings.HasPrefix(strings.TrimLeft(after, " \t"), "gpt-") ||
					strings.HasPrefix(strings.TrimLeft(after, " \t"), "gpt‑") {
					return false
				}
			}
		}
		// empty prompt line after the text is a strong submit signal
		if strings.Contains(plain, "›\n") || strings.Contains(plain, "› \n") ||
			strings.Contains(plain, "›\r") {
			return true
		}
	}
	return false
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
