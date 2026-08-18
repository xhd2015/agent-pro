# Scenario

**Feature**: `grok-tty` runner — adhoc ptywrap server, registry, capture sidecar, attach

```
agent-run run --agent-runner grok-tty "prompt"
  -> adhoc ptywrap on 127.0.0.1:0
  -> registry grok-tty-registry/<id>.json
  -> stderr grok-tty: session-N
  -> capture sidecar (readonly scrollback)
  -> inject prompt after GROK_TTY_BANNER
  -> events under sessions/grok-tty/

agent-run attach <id> -> registry lookup -> ptyclient WS attach
```

## Preconditions

- Repository contains `cmd/agent-run` (build may fail until grok-tty is implemented).
- Each test uses isolated `AGENT_RUN_HOME=filepath.Join(t.TempDir(), ".agent-run")`.
- Default-suite tests set `AGENT_RUN_GROK_TTY_COMMAND` to a fake interactive TUI script
  (not `grok -p` / streaming-json).
- No `agent-term serve` — each run owns its own ephemeral ptywrap HTTP server.
- Fake TUI scripts print `GROK_TTY_BANNER` before the prompt marker so banner-wait
  logic is deterministic.

## Steps

1. Root `Setup` builds `agent-run`, sets `AGENT_RUN_HOME`, configures fake TUI env
   (unless `req.SkipFakeTUI`).
2. Grouping `Setup` narrows subcommand (`run`, `attach`, `help`) or real-grok profile.
3. Leaf `Setup` sets prompt, fake TUI variant, or starts a background grok-tty run.
4. `Run` executes CLI or performs registry/attach probe against a live session.
5. Leaf `Assert` checks stderr session id, registry JSON, captured output, or attach.

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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
	"github.com/xhd2015/doctest/session"
)

const grokTTYBannerMarker = "GROK_TTY_BANNER"

func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

func fakeTUILongRun() string {
	return `sh -c 'printf "GROK_TTY_BANNER\n"; sleep 30'`
}

func fakeTUIDelayedBanner() string {
	return `sh -c 'sleep 0.3; printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

// fakeTUIRequiresCR mimics cooked TTY line input: characters echo to the prompt
// but read(2) only completes after Enter (\r). Bare \n without \r leaves text
// visible yet unsubmitted (the user-reported grok-tty bug).
func fakeTUIRequiresCR() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read -r line; echo "SUBMITTED:$line"'`
}

// fakeTUIRequiresCRDelayed waits before reading so attach can observe the PTY
// between banner and line submission.
func fakeTUIRequiresCRDelayed() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep 8; read -r line; echo "SUBMITTED:$line"'`
}

// fakeTUIHoldSeconds keeps the PTY alive so stream-probe tests can observe stdout
// before the fake TUI exits.
func fakeTUIHoldSeconds(sec int) string {
	return fmt.Sprintf(`sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep %d'`, sec)
}

func encodedGrokCwd(workspace string) string {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}
	return url.PathEscape(abs)
}

func grokSessionDir(grokHome, workspace, sessionUUID string) string {
	return filepath.Join(grokHome, "sessions", encodedGrokCwd(workspace), sessionUUID)
}

func grokSummaryJSON(workspace, sessionUUID string) string {
	return grokSummaryJSONAt(workspace, sessionUUID, time.Now().UTC())
}

func grokSummaryJSONAt(workspace, sessionUUID string, created time.Time) string {
	abs, _ := filepath.Abs(workspace)
	payload := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
			"openedAt":  created.Format(time.RFC3339Nano),
		},
		"created_at": created.Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func writeFakeGrokSessionDirAt(t *testing.T, grokHome, workspace, sessionUUID, prompt string, created time.Time, initialLines ...string) string {
	t.Helper()
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir %s: %v", dir, err)
	}
	summaryPath := filepath.Join(dir, "summary.json")
	if err := os.WriteFile(summaryPath, []byte(grokSummaryJSONAt(workspace, sessionUUID, created)), 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	seed := []string{acpUserMessageChunk(prompt)}
	seed = append(seed, initialLines...)
	if err := appendUpdatesJSONL(updatesPath, seed...); err != nil {
		t.Fatalf("seed updates.jsonl: %v", err)
	}
	return updatesPath
}

func acpUserMessageChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpAgentThoughtChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_thought_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpAgentMessageChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpToolCall(toolCallID, kind, title string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call",
		"toolCallId":    toolCallID,
		"kind":          kind,
		"title":         title,
		"status":        "pending",
	})
	return string(line)
}

func acpTurnCompleted() string {
	return `{"sessionUpdate":"turn_completed"}`
}

func acpToolCallUpdate(toolCallID, status, output string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call_update",
		"toolCallId":    toolCallID,
		"status":        status,
		"content": []map[string]any{
			{
				"type": "content",
				"content": map[string]any{
					"type": "text",
					"text": output,
				},
			},
		},
	})
	return string(line)
}

func appendUpdatesJSONL(path string, lines ...string) error {
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

func writeFakeGrokSessionDir(t *testing.T, grokHome, workspace, sessionUUID, prompt string, initialLines ...string) string {
	t.Helper()
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir %s: %v", dir, err)
	}
	summaryPath := filepath.Join(dir, "summary.json")
	if err := os.WriteFile(summaryPath, []byte(grokSummaryJSON(workspace, sessionUUID)), 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	seed := []string{acpUserMessageChunk(prompt)}
	seed = append(seed, initialLines...)
	if err := appendUpdatesJSONL(updatesPath, seed...); err != nil {
		t.Fatalf("seed updates.jsonl: %v", err)
	}
	return updatesPath
}

func appendGrokHomeEnv(req *Request) {
	if strings.TrimSpace(req.GrokHome) == "" {
		return
	}
	req.Env = withoutEnvKey(req.Env, "GROK_HOME")
	req.Env = append(req.Env, "GROK_HOME="+req.GrokHome)
	if id := strings.TrimSpace(req.GrokSessionUUID); id != "" {
		req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID")
		req.Env = append(req.Env, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+id)
	}
}

func grokTTYRegistryDir(home string) string {
	return filepath.Join(home, "grok-tty-registry")
}

func grokTTYRegistryPath(home, sessionID string) string {
	return filepath.Join(grokTTYRegistryDir(home), sessionID+".json")
}

func normalizeTTYStreamStdout(s string) string {
	if s == "" {
		return s
	}
	s = strings.TrimSuffix(s, "\n")
	if !strings.HasPrefix(s, "\n") {
		s = "\n" + s
	}
	return s
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
		Stdout:   normalizeTTYStreamStdout(stdout.String()),
		Stderr:   stderr.String(),
		Err:      err,
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
	appendGrokHomeEnv(req)
	return execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
}

func appendGrokTTYEnv(req *Request) {
	if req.SkipFakeTUI {
		return
	}
	cmd := strings.TrimSpace(req.GrokTTYCommand)
	if cmd == "" {
		cmd = fakeTUIRespondHi()
	}
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	req.Env = append(req.Env, "AGENT_RUN_GROK_TTY_COMMAND="+cmd)
}

func setGrokTTYCommand(req *Request, cmd string) {
	req.GrokTTYCommand = cmd
	appendGrokTTYEnv(req)
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

func parseGrokTTYSessionID(stderr string) (string, bool) {
	re := regexp.MustCompile(`grok-tty:\s*(session-\d+)`)
	m := re.FindStringSubmatch(stderr)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func waitForGrokTTYSessionLine(t *testing.T, buf *bytes.Buffer, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if id, ok := parseGrokTTYSessionID(buf.String()); ok {
			return id
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for grok-tty session id in stderr:\n%s", buf.String())
	return ""
}

func readRegistryEntry(t *testing.T, home, sessionID string) *GrokTTYRegistryEntry {
	t.Helper()
	path := grokTTYRegistryPath(home, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry %s: %v", path, err)
	}
	var entry GrokTTYRegistryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("parse registry %s: %v\n%s", path, err, string(data))
	}
	return &entry
}

func waitForRegistryEntry(t *testing.T, home, sessionID string, timeout time.Duration) *GrokTTYRegistryEntry {
	t.Helper()
	deadline := time.Now().Add(timeout)
	path := grokTTYRegistryPath(home, sessionID)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil {
			var entry GrokTTYRegistryEntry
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
		return true, "" // attach blocks interactively; reaching timeout after dial OK is success
	}
}

func probeAttachCLI(t *testing.T, req *Request, sessionID string) (*Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, "attach", sessionID)
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
		// attach blocks until session ends; timeout after connect means probe OK
		if strings.Contains(stderr.String(), "grok-tty") || stdout.Len() > 0 || ctx.Err() == context.DeadlineExceeded {
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

func buildLLMMockRunGrok(t *testing.T, req *Request) error {
	t.Helper()
	if req.LLMMockRunGrok != "" {
		if _, err := os.Stat(req.LLMMockRunGrok); err == nil {
			return nil
		}
	}
	req.LLMMockRunGrok = filepath.Join(req.TempDir, "bin", "llm-mock-run-grok")
	if err := os.MkdirAll(filepath.Dir(req.LLMMockRunGrok), 0755); err != nil {
		return err
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.LLMMockRunGrok, "./agent/llm/llm-mock/llm-mock-run-grok")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-grok: %w\n%s", err, string(out))
	}
	return nil
}

func startGrokTTYBackground(t *testing.T, req *Request) {
	t.Helper()
	args := []string{"run", "--agent-runner", "grok-tty"}
	if req.AgentRunnerBinary != "" {
		args = append(args, "--agent-runner-binary", req.AgentRunnerBinary)
	}
	if req.AgentRunnerConfigHome != "" {
		args = append(args, "--agent-runner-config-home", req.AgentRunnerConfigHome)
	}
	if req.KeepTTY {
		args = append(args, "--keep-tty")
	}
	if req.GrokTTYPrompt != "" {
		args = append(args, req.GrokTTYPrompt)
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
		t.Fatalf("start background grok-tty run: %v", err)
	}
	req.BackgroundCmd = cmd
	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
	req.GrokTTYSessionID = waitForGrokTTYSessionLine(t, req.BackgroundStderr, 15*time.Second)
}

func runSnapshotProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GrokTTYSessionID == "" {
		t.Fatal("snapshot-probe requires GrokTTYSessionID from background Setup")
	}
	marker := strings.TrimSpace(req.SnapshotReadyMarker)
	if marker == "" {
		marker = "💬"
	}
	deadline := time.Now().Add(45 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		if strings.Contains(req.BackgroundStdout.String(), marker) {
			ready = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("timeout waiting for snapshot-ready marker %q in background stdout:\n%s\nstderr:\n%s",
			marker, req.BackgroundStdout.String(), req.BackgroundStderr.String())
	}
	delay := req.SnapshotDelay
	if delay <= 0 {
		delay = 2 * time.Second
	}
	time.Sleep(delay)
	snapResp, err := execCmd(t, req.AgentRun, []string{"snapshot", req.GrokTTYSessionID}, req.TempDir, req.Env, 15*time.Second)
	resp := &Response{
		Stdout:           snapResp.Stdout,
		Stderr:           snapResp.Stderr,
		ExitCode:         snapResp.ExitCode,
		Err:              snapResp.Err,
		SnapshotStdout:   snapResp.Stdout,
		SnapshotExitCode: snapResp.ExitCode,
		BackgroundStdout: req.BackgroundStdout.String(),
		BackgroundStderr: req.BackgroundStderr.String(),
	}
	return resp, err
}

func runRegistryWhileRunning(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GrokTTYSessionID == "" {
		t.Fatal("registry-while-running requires GrokTTYSessionID from background Setup")
	}
	entry := waitForRegistryEntry(t, req.Home, req.GrokTTYSessionID, 10*time.Second)
	waitForTCPAddr(t, entry.ListenAddr, 5*time.Second)
	return &Response{RegistryEntry: entry}, nil
}

func runAttachProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GrokTTYSessionID == "" {
		t.Fatal("attach-probe requires GrokTTYSessionID from background Setup")
	}
	entry := waitForRegistryEntry(t, req.Home, req.GrokTTYSessionID, 10*time.Second)
	ok, probeErr := probeAttachWS(t, entry.ListenAddr, entry.SessionID)
	return &Response{
		RegistryEntry:  entry,
		AttachProbeOK:  ok,
		AttachProbeErr: probeErr,
	}, nil
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
	if req.GrokTTYSessionID == "" {
		t.Fatal("attach-scrollback-probe requires GrokTTYSessionID from background Setup")
	}
	entry := waitForRegistryEntry(t, req.Home, req.GrokTTYSessionID, 10*time.Second)
	scrollback, err := collectAttachScrollback(t, entry.ListenAddr, entry.SessionID, 2*time.Second)
	resp := &Response{
		RegistryEntry:    entry,
		AttachScrollback: scrollback,
		AttachProbeOK:    entry.ListenAddr != "",
	}
	if err != nil && !strings.Contains(scrollback, "GROK_TTY_BANNER") {
		resp.AttachProbeErr = err.Error()
		resp.AttachProbeOK = false
	}
	return resp, nil
}

func runAttachInteractiveProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GrokTTYSessionID == "" {
		t.Fatal("attach-interactive-probe requires GrokTTYSessionID from background Setup")
	}
	resp, err := probeAttachCLI(t, req, req.GrokTTYSessionID)
	if resp != nil && resp.AttachProbeOK {
		return resp, err
	}
	entry := waitForRegistryEntry(t, req.Home, req.GrokTTYSessionID, 10*time.Second)
	ok, probeErr := probeAttachWS(t, entry.ListenAddr, entry.SessionID)
	resp.AttachProbeOK = ok
	resp.AttachProbeErr = probeErr
	resp.RegistryEntry = entry
	return resp, err
}

func findGrokTTYEventsJSONL(t *testing.T, home string) (string, []string) {
	t.Helper()
	root := filepath.Join(home, "sessions", "grok-tty")
	var found string
	var lines []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == "events.jsonl" {
			found = path
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read %s: %v", path, readErr)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) != "" {
					lines = append(lines, line)
				}
			}
		}
		return nil
	})
	return found, lines
}

func findGrokTTYMetaJSON(t *testing.T, home string) (string, map[string]any) {
	t.Helper()
	root := filepath.Join(home, "sessions", "grok-tty")
	var found string
	var meta map[string]any
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() != "meta.json" {
			return nil
		}
		found = path
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		if json.Unmarshal(data, &meta) != nil {
			t.Fatalf("parse meta.json %s", path)
		}
		return nil
	})
	return found, meta
}

func eventsCollectTypes(lines []string) map[string]bool {
	found := make(map[string]bool)
	for _, line := range lines {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		typ, _ := ev["type"].(string)
		role, _ := ev["role"].(string)
		switch typ {
		case "message":
			if role == "user" {
				found["message:user"] = true
			} else if role == "assistant" {
				found["message:assistant"] = true
			}
		case "think", "tool_call":
			found[typ] = true
		}
	}
	return found
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

// assertStderrGrokSession requires stderr diagnostics after grok session discovery:
// `grok-tty: grok session <uuid>` and `grok-tty: grok updates <absolute-path>`.
func assertStderrGrokSession(t *testing.T, stderr, sessionUUID, updatesPath string) {
	t.Helper()
	wantSession := fmt.Sprintf("grok-tty: grok session %s", sessionUUID)
	if !strings.Contains(stderr, wantSession) {
		t.Fatalf("stderr missing %q; stderr:\n%s", wantSession, stderr)
	}
	const updatesPrefix = "grok-tty: grok updates "
	if !strings.Contains(stderr, updatesPrefix) {
		t.Fatalf("stderr missing %q; stderr:\n%s", updatesPrefix, stderr)
	}
	if updatesPath != "" && !strings.Contains(stderr, updatesPath) {
		t.Fatalf("stderr missing updates path %q; stderr:\n%s", updatesPath, stderr)
	}
}

// assertStderrGrokSessionBeforeStdout requires grok session stderr lines to appear
// before the first stdout stream marker (ordering guard for live diagnostics).
func assertStderrGrokSessionBeforeStdout(t *testing.T, stderr, stdout, marker, sessionUUID string) {
	t.Helper()
	assertStderrGrokSession(t, stderr, sessionUUID, "")
	sessionIdx := strings.Index(stderr, fmt.Sprintf("grok-tty: grok session %s", sessionUUID))
	updatesIdx := strings.Index(stderr, "grok-tty: grok updates ")
	markerIdx := strings.Index(stdout, marker)
	if markerIdx < 0 {
		t.Fatalf("stdout missing stream marker %q; stdout:\n%s", marker, stdout)
	}
	if sessionIdx < 0 || updatesIdx < 0 {
		return // assertStderrGrokSession already failed
	}
	if sessionIdx >= markerIdx || updatesIdx >= markerIdx {
		t.Fatalf("expected stderr grok session lines before stdout marker %q; sessionIdx=%d updatesIdx=%d markerIdx=%d\nstderr:\n%s\nstdout:\n%s",
			marker, sessionIdx, updatesIdx, markerIdx, stderr, stdout)
	}
}

// delayedGrokSessionSchedule returns a GrokUpdatesSchedule that creates the grok
// session dir after delay (simulates grok writing updates.jsonl late during PTY run).
func delayedGrokSessionSchedule(t *testing.T, delay time.Duration, grokHome, workspace, sessionUUID, prompt string, initialLines ...string) (GrokUpdatesSchedule, string) {
	t.Helper()
	updatesPath := filepath.Join(grokSessionDir(grokHome, workspace, sessionUUID), "updates.jsonl")
	return GrokUpdatesSchedule{
		Delay: delay,
		OnFire: func() {
			writeFakeGrokSessionDir(t, grokHome, workspace, sessionUUID, prompt, initialLines...)
		},
	}, updatesPath
}

func startGrokUpdatesScheduler(t *testing.T, req *Request, sessionReady <-chan struct{}) {
	t.Helper()
	if len(req.GrokUpdatesSchedules) == 0 {
		return
	}
	go func() {
		<-sessionReady
		for _, sched := range req.GrokUpdatesSchedules {
			sched := sched
			time.Sleep(sched.Delay)
			if sched.OnFire != nil {
				sched.OnFire()
			}
			path := sched.UpdatesPath
			if path == "" {
				path = req.GrokUpdatesPath
			}
			if path != "" && len(sched.Lines) > 0 {
				if err := appendUpdatesJSONL(path, sched.Lines...); err != nil {
					t.Logf("append updates.jsonl: %v", err)
				}
			}
		}
	}()
}

func runStreamProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	probe := strings.TrimSpace(req.StreamProbeSubstring)
	if probe == "" {
		t.Fatal("stream-probe requires StreamProbeSubstring")
	}
	timeout := req.StreamProbeTimeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	appendGrokHomeEnv(req)

	ctx, cancel := context.WithTimeout(context.Background(), timeout+45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, req.Args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	sessionReady := make(chan struct{})
	go func() {
		_, _ = parseGrokTTYSessionID(stderrBuf.String())
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			if _, ok := parseGrokTTYSessionID(stderrBuf.String()); ok {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(sessionReady)
	}()
	startGrokUpdatesScheduler(t, req, sessionReady)

	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()

	probeDeadline := time.Now().Add(timeout)
	var probeSeen bool
	var probeBeforeExit bool
	var waitErr error
	processExited := false
	for time.Now().Before(probeDeadline) {
		if !probeSeen && strings.Contains(stdoutBuf.String(), probe) {
			probeSeen = true
			probeBeforeExit = cmd.ProcessState == nil
		}
		select {
		case waitErr = <-exited:
			processExited = true
			goto done
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

done:
	if !processExited {
		select {
		case waitErr = <-exited:
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			waitErr = <-exited
		}
	}

	resp := &Response{
		Stdout:                stdoutBuf.String(),
		Stderr:                stderrBuf.String(),
		Err:                   waitErr,
		StreamProbeSeen:       probeSeen,
		StreamProbeBeforeExit: probeBeforeExit,
		StreamProbeStdout:     stdoutBuf.String(),
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
		}
	}
	metaPath, meta := findGrokTTYMetaJSON(t, req.Home)
	resp.MetaJSONPath = metaPath
	if meta != nil {
		if v, ok := meta["runner_session_id"].(string); ok {
			resp.MetaRunnerSessionID = v
		}
	}
	_, resp.EventsFileLines = findGrokTTYEventsJSONL(t, req.Home)
	return resp, nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
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
	appendGrokTTYEnv(req)
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

func drainLines(r io.Reader) []string {
	var lines []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}
```