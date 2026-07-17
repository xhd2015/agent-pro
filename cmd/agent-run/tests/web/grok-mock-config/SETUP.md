# Scenario

**Feature**: web grok mock config flags apply to grok-tty web sessions

```
agent-run web --grok-home PATH --grok-tty-runner-binary SPEC --agent-runner grok-tty
POST /api/agent-run/sessions {runner: grok-tty}
  -> child uses SPEC under GROK_HOME=PATH
  -> grok session tree under PATH
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Web server uses `--token test` and `--port 0`.
- `LLM_MOCK_RUN_GROK_COMMAND` fake hook seeds `updates.jsonl` under `$GROK_HOME`.

## Steps

1. Root `Setup` builds `agent-run` and `llm-mock-run-grok`, writes argv-probe wrapper script.
2. Leaf `Setup` starts web with grok flags, POSTs grok-tty session, waits for finish.
3. `Run` reads argv probe file and session detail.
4. `Assert` checks grok home session tree and mock binary invocation.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

const webGrokMockUUID = "c3333333-3333-4333-8333-333333333333"

func writeMockWrapper(t *testing.T, wrapperPath, probePath, prompt, sessionUUID, marker string) error {
	t.Helper()
	script := fmt.Sprintf(`#!/bin/sh
echo "MOCK_WRAPPER_INVOKED=1" >> %q
echo "ARGV_RECORD=$*" >> %q
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
submitted="${line:-%s}"
wd=$(pwd)
enc=$(python3 -c 'import os,sys,urllib.parse
p=os.path.abspath(sys.argv[1])
if p.startswith("/private/var/"): p="/var/"+p[len("/private/var/"):]
elif p.startswith("/private/tmp/"): p="/tmp/"+p[len("/private/tmp/"):]
print(urllib.parse.quote(p, safe=""))' "$wd")
dir="$GROK_HOME/sessions/$enc/%s"
mkdir -p "$dir"
now=$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)
cat > "$dir/summary.json" <<EOF
{"info":{"cwd":"$wd","sessionId":"%s","openedAt":"$now"},"created_at":"$now"}
EOF
cat > "$dir/updates.jsonl" <<EOF
{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"$submitted"}}
{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"%s"}}
EOF
sleep 2
exit 0
`, probePath, probePath, prompt, sessionUUID, sessionUUID, marker)
	return os.WriteFile(wrapperPath, []byte(script), 0755)
}

func llmMockWebGrokHook(prompt, sessionUUID, marker string) string {
	return fmt.Sprintf(`sh -c '
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
submitted="${line:-%s}"
wd=$(pwd)
enc=$(python3 -c 'import os,sys,urllib.parse
p=os.path.abspath(sys.argv[1])
if p.startswith("/private/var/"): p="/var/"+p[len("/private/var/"):]
elif p.startswith("/private/tmp/"): p="/tmp/"+p[len("/private/tmp/"):]
print(urllib.parse.quote(p, safe=""))' "$wd")
dir="$GROK_HOME/sessions/$enc/%s"
mkdir -p "$dir"
now=$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)
cat > "$dir/summary.json" <<EOF
{"info":{"cwd":"$wd","sessionId":"%s","openedAt":"$now"},"created_at":"$now"}
EOF
cat > "$dir/updates.jsonl" <<EOF
{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"$submitted"}
{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"%s"}}
EOF
exit 0
'`, prompt, sessionUUID, sessionUUID, marker)
}

func parseWebListenURL(stderr string) (string, bool) {
	re := regexp.MustCompile(`https?://127\.0\.0\.1:(\d+)`)
	m := re.FindStringSubmatch(stderr)
	if len(m) < 2 {
		return "", false
	}
	return "http://127.0.0.1:" + m[1], true
}

func waitForListenPort(stderr *bytes.Buffer, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if u, ok := parseWebListenURL(stderr.String()); ok {
			return u, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for listen URL")
}

func httpPostJSON(t *testing.T, url, bearer, jsonBody string) (int, string) {
	t.Helper()
	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if bearer != "" {
		httpReq.Header.Set("Authorization", "Bearer "+bearer)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("http POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func httpGet(t *testing.T, url, bearer string) (int, string) {
	t.Helper()
	client := &http.Client{Timeout: 30 * time.Second}
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if bearer != "" {
		httpReq.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("http GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func waitForHealth(t *testing.T, baseURL, bearer string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/health"
	for time.Now().Before(deadline) {
		status, _ := httpGet(t, url, bearer)
		if status == 200 {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func postCreateSession(t *testing.T, baseURL, bearer, runner, prompt string) string {
	payload, _ := json.Marshal(map[string]string{"runner": runner, "prompt": prompt})
	status, body := httpPostJSON(t, baseURL+"/api/agent-run/sessions", bearer, string(payload))
	if status != http.StatusAccepted && status != 200 {
		t.Fatalf("POST sessions: status=%d body=%q", status, body)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("parse create response: %v", err)
	}
	sess, _ := parsed["session"].(map[string]any)
	id, _ := sess["session_id"].(string)
	if strings.TrimSpace(id) == "" {
		t.Fatalf("empty session_id: %q", body)
	}
	return id
}

func sessionStatus(detailJSON string) string {
	var parsed map[string]any
	if json.Unmarshal([]byte(detailJSON), &parsed) != nil {
		return ""
	}
	sess, _ := parsed["session"].(map[string]any)
	status, _ := sess["status"].(string)
	return status
}

func waitForSessionFinished(t *testing.T, baseURL, bearer, runner, sessionID string, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("%s/api/agent-run/sessions/%s/%s", baseURL, runner, sessionID)
	for time.Now().Before(deadline) {
		_, body := httpGet(t, url, bearer)
		if sessionStatus(body) == "finished" {
			return body
		}
		time.Sleep(100 * time.Millisecond)
	}
	_, body := httpGet(t, url, bearer)
	t.Fatalf("timeout waiting for finished status, got %q: %s", sessionStatus(body), body)
	return body
}

func canonicalWorkspacePath(path string) string {
	path = filepath.Clean(path)
	if abs, err := filepath.Abs(path); err == nil {
		path = filepath.Clean(abs)
	}
	if strings.HasPrefix(path, "/private/") {
		alt := strings.TrimPrefix(path, "/private")
		if alt != "" && alt[0] == '/' {
			if a, errA := os.Lstat(path); errA == nil {
				if b, errB := os.Lstat(alt); errB == nil && os.SameFile(a, b) {
					return filepath.Clean(alt)
				}
			}
		}
	}
	return path
}

func grokSessionUpdatesExists(grokHome, workspace, sessionUUID string) bool {
	enc := url.PathEscape(canonicalWorkspacePath(workspace))
	updatesPath := filepath.Join(grokHome, "sessions", enc, sessionUUID, "updates.jsonl")
	if _, err := os.Stat(updatesPath); err == nil {
		return true
	}
	var found bool
	_ = filepath.Walk(filepath.Join(grokHome, "sessions"), func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Name() == "updates.jsonl" && strings.Contains(p, sessionUUID) {
			found = true
		}
		return nil
	})
	return found
}

func startWebWithGrokMock(t *testing.T, req *Request) {
	t.Helper()
	args := []string{
		"web", "--no-open", "--token", req.WebToken, "--port", "0",
		"--agent-runner", "grok-tty",
		"--grok-home", req.GrokHome,
		"--grok-tty-runner-binary", req.GrokTTYRunnerBinary,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	req.webProcessStderr = &bytes.Buffer{}
	cmd.Stderr = req.webProcessStderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start web: %v", err)
	}
	req.WebCmd = cmd
	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
	baseURL, err := waitForListenPort(req.webProcessStderr, 15*time.Second)
	if err != nil {
		t.Fatalf("web listen: %v\nstderr:\n%s", err, req.webProcessStderr.String())
	}
	req.WebBaseURL = strings.TrimRight(baseURL, "/")
	if !waitForHealth(t, req.WebBaseURL, req.WebToken, 15*time.Second) {
		t.Fatalf("health failed; stderr:\n%s", req.webProcessStderr.String())
	}
}

func runWebGrokMockProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}
	if req.WebBaseURL == "" || req.SessionID == "" {
		return resp, fmt.Errorf("web session not initialized")
	}
	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(req.ArgvProbePath); err == nil {
			resp.ArgvProbe = string(data)
		}
		if strings.Contains(resp.ArgvProbe, "MOCK_WRAPPER_INVOKED=1") &&
			grokSessionUpdatesExists(req.GrokHome, req.TempDir, webGrokMockUUID) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	status, body := httpGet(t, fmt.Sprintf("%s/api/agent-run/sessions/%s/%s", req.WebBaseURL, req.SessionRunner, req.SessionID), req.WebToken)
	resp.HTTPStatus = status
	resp.HTTPBody = body
	return resp, nil
}

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	req.LLMMockRunGrok = filepath.Join(req.TempDir, "bin", "llm-mock-run-grok")
	req.MockWrapperPath = filepath.Join(req.TempDir, "bin", "mock-wrapper.sh")
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe.log")
	req.GrokHome = filepath.Join(req.TempDir, "web-grok-home")
	req.WebToken = "test"
	req.SessionRunner = "grok-tty"
	req.CreatePrompt = "web mock config probe"
	req.GrokSessionUUID = webGrokMockUUID

	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	build := exec.Command("go", "build", "-buildvcs=false", "-o", req.AgentRun, "./cmd/agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	buildMock := exec.Command("go", "build", "-buildvcs=false", "-o", req.LLMMockRunGrok, "./agent/llm/llm-mock/llm-mock-run-grok")
	buildMock.Dir = req.RepoRoot
	if out, err := buildMock.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-grok: %w\n%s", err, string(out))
	}
	if err := writeMockWrapper(t, req.MockWrapperPath, req.ArgvProbePath, req.CreatePrompt, webGrokMockUUID, "WEB_MOCK_STREAM_MARKER"); err != nil {
		return err
	}
	req.GrokTTYRunnerBinary = req.MockWrapperPath
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+webGrokMockUUID,
	)
	return nil
}
```