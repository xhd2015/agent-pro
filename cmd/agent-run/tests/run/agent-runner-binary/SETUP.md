# Scenario

**Feature**: `agent-run run --agent-runner-binary SPEC` drives grok-tty PTY argv

```
agent-run run --agent-runner grok-tty --agent-runner-binary SPEC "prompt"
  -> (no AGENT_RUN_GROK_TTY_COMMAND)
  -> PTY spawns resolved binary + user flags + grok-tty defaults
  -> fake script stderr/scrollback records argv
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Each test uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Grouping `Setup` unsets `AGENT_RUN_GROK_TTY_COMMAND` so `--agent-runner-binary` is exercised.
- Fake runner scripts print `GROK_TTY_BANNER` before reading stdin.

## Steps

1. Root `Setup` builds `agent-run` (and `llm-mock-run-grok` when needed).
2. Grouping `Setup` prefixes `run --agent-runner grok-tty` and clears the grok-tty hook env.
3. Leaf `Setup` writes fake runner scripts, sets `--agent-runner-binary`, and prompt.
4. `Run` executes blocking `agent-run run`.
5. Leaf `Assert` checks stderr argv record, model precedence, or grok session discovery.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const grokTTYBannerMarker = "GROK_TTY_BANNER"

func writeArgvRecordingRunner(t *testing.T, dir, name, probePath string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	script := fmt.Sprintf(`#!/bin/sh
echo "ARGV_RECORD=$*" > %q
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
echo "Response: ${line:-done}"
exit 0
`, probePath)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write argv runner %s: %v", path, err)
	}
	return path
}

func writeModelProbeRunner(t *testing.T, dir string, innerModel, probePath string) string {
	t.Helper()
	path := filepath.Join(dir, "model-probe.sh")
	script := fmt.Sprintf(`#!/bin/sh
echo "ARGV_RECORD=$*" > %q
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
echo "MODEL_PROBE:%s"
exit 0
`, probePath, innerModel)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write model probe runner: %v", err)
	}
	return path
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

func acpUserMessageChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
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

func grokSummaryJSON(workspace, sessionUUID string) string {
	abs, _ := filepath.Abs(workspace)
	payload := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
			"openedAt":  time.Now().UTC().Format(time.RFC3339Nano),
		},
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func writeFakeGrokSessionDir(t *testing.T, grokHome, workspace, sessionUUID, prompt string, initialLines ...string) string {
	t.Helper()
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir %s: %v", dir, err)
	}
	summaryPath := filepath.Join(dir, "summary.json")
	abs, _ := filepath.Abs(workspace)
	payload := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
			"openedAt":  time.Now().UTC().Format(time.RFC3339Nano),
		},
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(payload)
	if err := os.WriteFile(summaryPath, b, 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	seed := []string{acpUserMessageChunk(prompt)}
	seed = append(seed, initialLines...)
	f, err := os.OpenFile(updatesPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open updates.jsonl: %v", err)
	}
	defer f.Close()
	for _, line := range seed {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := fmt.Fprintln(f, line); err != nil {
			t.Fatalf("append updates.jsonl: %v", err)
		}
	}
	return updatesPath
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

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 60 * time.Second
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
	return execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
}

func writeLLMMockRunGrokHarness(t *testing.T, path, prompt, sessionUUID, assistantMarker, grokHomeProbe string) error {
	t.Helper()
	script := fmt.Sprintf(`#!/bin/sh
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
prompt="${line:-%s}"
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
{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"$prompt"}}
{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"%s"}}
EOF
echo "$GROK_HOME" > %q
find "$GROK_HOME" -name updates.jsonl 2>/dev/null >> %q
sleep 5
exit 0
`, prompt, sessionUUID, sessionUUID, assistantMarker, grokHomeProbe, grokHomeProbe)
	return os.WriteFile(path, []byte(script), 0755)
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
	build := exec.Command("go", "build", "-buildvcs=false", "-o", req.LLMMockRunGrok, "./agent/llm/llm-mock/llm-mock-run-grok")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-grok: %w\n%s", err, string(out))
	}
	return nil
}

func llmMockGrokHookCreatesSession(prompt, sessionUUID, assistantMarker string) string {
	return fmt.Sprintf(`sh -c '
printf "GROK_TTY_BANNER\nGrok › "
wd=$(pwd)
enc=$(python3 -c 'import os,sys,urllib.parse
p=os.path.abspath(sys.argv[1])
if p.startswith("/private/var/"): p="/var/"+p[len("/private/var/"):]
elif p.startswith("/private/tmp/"): p="/tmp/"+p[len("/private/tmp/"):]
print(urllib.parse.quote(p, safe=""))' "$wd")
dir="$GROK_HOME/sessions/$enc/%s"
mkdir -p "$dir"
cat > "$dir/summary.json" <<EOF
{"info":{"cwd":"$wd","sessionId":"%s"},"created_at":"now"}
EOF
cat > "$dir/updates.jsonl" <<EOF
{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"%s"}}
{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"%s"}}
EOF
read -r line || true
exit 0
'`, sessionUUID, sessionUUID, prompt, assistantMarker)
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertStderrGrokSession(t *testing.T, stderr, sessionUUID string) {
	t.Helper()
	want := fmt.Sprintf("grok-tty: grok session %s", sessionUUID)
	if !strings.Contains(stderr, want) {
		t.Fatalf("stderr missing %q; stderr:\n%s", want, stderr)
	}
	if !strings.Contains(stderr, "grok-tty: grok updates ") {
		t.Fatalf("stderr missing grok updates path; stderr:\n%s", stderr)
	}
}

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command("go", "build", "-buildvcs=false", "-o", req.AgentRun, "./cmd/agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	req.Args = []string{"run", "--agent-runner", "grok-tty"}
	return nil
}
```