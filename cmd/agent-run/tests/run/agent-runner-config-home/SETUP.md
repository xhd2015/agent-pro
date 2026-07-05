# Scenario

**Feature**: `--agent-runner-config-home` drives discovery and child `GROK_HOME`

```
agent-run run --agent-runner-config-home PATH --agent-runner-binary <fake>
  -> discovery reads PATH/sessions/...
  -> PTY argv: env GROK_HOME=PATH <binary> ...
```

## Preconditions

- Repository contains `cmd/agent-run`.
- `AGENT_RUN_GROK_TTY_COMMAND` is unset; fake runner scripts replace `grok`.
- Each test uses isolated `AGENT_RUN_HOME` and temp config home under `t.TempDir()`.

## Steps

1. Root `Setup` builds `agent-run`, clears grok-tty hook env.
2. Grouping `Setup` prefixes `run --agent-runner grok-tty`.
3. Leaf `Setup` seeds grok session dir or env-logging runner + config home flag.
4. `Run` executes blocking `agent-run run`.
5. `Assert` checks stderr grok session lines or child env dump.

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

func writeHoldFakeRunner(t *testing.T, dir, name string, holdSec int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	script := fmt.Sprintf(`#!/bin/sh
printf "GROK_TTY_BANNER\nGrok › "
sleep %d
exit 0
`, holdSec)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write hold runner: %v", err)
	}
	return path
}

func writeEnvLoggingRunner(t *testing.T, dir, probePath string) string {
	t.Helper()
	path := filepath.Join(dir, "env-logger.sh")
	script := fmt.Sprintf(`#!/bin/sh
env | grep -E '^(GROK_HOME|AGENT_RUNNER_CONFIG_HOME)=' | sort > %q
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
echo "ENV_LOGGER_OK"
exit 0
`, probePath)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write env logger: %v", err)
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

func writeFakeGrokSessionDir(t *testing.T, grokHome, workspace, sessionUUID, prompt string, initialLines ...string) string {
	t.Helper()
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir %s: %v", dir, err)
	}
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
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
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

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertStderrGrokSession(t *testing.T, stderr, sessionUUID, updatesPath string) {
	t.Helper()
	wantSession := fmt.Sprintf("grok-tty: grok session %s", sessionUUID)
	if !strings.Contains(stderr, wantSession) {
		t.Fatalf("stderr missing %q; stderr:\n%s", wantSession, stderr)
	}
	if !strings.Contains(stderr, "grok-tty: grok updates ") {
		t.Fatalf("stderr missing grok updates path; stderr:\n%s", stderr)
	}
	if updatesPath != "" && !strings.Contains(stderr, updatesPath) {
		t.Fatalf("stderr missing updates path %q; stderr:\n%s", updatesPath, stderr)
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
	build := exec.Command("go", "build", "-o", req.AgentRun, "./cmd/agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	req.Env = withoutEnvKey(req.Env, "GROK_HOME")
	req.Args = []string{"run", "--agent-runner", "grok-tty"}
	return nil
}
```