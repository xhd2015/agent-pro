# Scenario

**Feature**: `llm-mock run opencode` orchestrator — mock server + isolated opencode env + opencode foreground

```
# resolve config
orchestrator -> mockconfig loader -> JSON config

# background mock, foreground opencode
orchestrator -> llm-mock HTTP server (background)
orchestrator -> set OPENCODE_CONFIG_DIR + OPENCODE_CONFIG_CONTENT (baseURL -> mock, openai-compatible)
orchestrator -> opencode CLI (foreground, fake or real)

# teardown
opencode exit <- orchestrator tears down mock (no session mirror poll)
integration <- output Paris from mocked Chat Completions API
```

## Preconditions

- Repository contains `agent/llm/llm-mock` and (post-implementation) `agent/llm/llm-mock/llm-mock-run-opencode`.
- Root `Setup` builds `llm-mock` and `llm-mock-run-opencode` when the shortcut package exists.
- Default plumbing tests set `LLM_MOCK_RUN_OPENCODE_COMMAND` to a fake opencode shell hook.
- `integration/` leaves require real `opencode` on PATH (`label: real-opencode, slow`).

## Steps

1. Root `Setup` resolves `RepoRoot`, builds `llm-mock` and optionally `llm-mock-run-opencode`.
2. Grouping `Setup` narrows config/home/CLI/integration profile.
3. Leaf `Setup` sets `Request` fields (config JSON, env mode, fake opencode, opencode args).
4. `Run` writes temp config files, sets env, executes orchestrator entrypoint.
5. Leaf `Assert` checks exit code, stdout/stderr, opencode config env, or log JSONL files.

## Context

- `Request.ConfigEnv` — `"file"` (`LLM_MOCK_CONFIG_FILE`) or `""` (neither).
- `Request.FakeOpencodeCmd` — replaces opencode binary for plumbing tests.
- `Request.UseShortcut` — `true` runs `llm-mock-run-opencode` instead of `llm-mock run opencode`.
- `Response.OpencodeConfigDirUsed` — parsed from fake opencode stdout/stderr or explicit `LLM_MOCK_OPENCODE_CONFIG_DIR`.
- Fake opencode curls read `baseURL` from `$OPENCODE_CONFIG_CONTENT` JSON (not a config file on disk).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const fakeOpencodePrintConfigDir = `sh -c 'echo OPENCODE_CONFIG_DIR=$OPENCODE_CONFIG_DIR; exit 0'`

const fakeOpencodeCurlChatCompletionsOnce = `sh -c '
base=$(printf "%s" "$OPENCODE_CONFIG_CONTENT" | sed -n "s/.*\"baseURL\"[[:space:]]*:[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" | head -1)
body=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -H "Authorization: Bearer sk-mock" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"config-only-prompt\"}]}")
echo "$body"
'`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = assertSuccess
	_ = assertContains
	_ = assertNotContains
	_ = assertExitCode
	_ = envWithOverrides
	_ = parseOpencodeConfigDirFromOutput
	_ = parseOpencodeHomeFromOutput
	_ = parseOpencodeConfigContentFromEnv
	_ = minimalMockConfigJSON
	_ = fakeOpencodePrintConfigDir
	_ = fakeOpencodeCurlChatCompletionsOnce
	_ = readJSONLLinesFromContent
	_ = parseAgentEventMaps
	_ = parseHTTPExchangeMaps
	_ = installFakeOpencodeEchoArgv
	_ = agentEventsHaveTypes
	_ = httpLogHasChatCompletionsModel

	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}

	tmp := t.TempDir()
	req.BinaryPath = filepath.Join(tmp, "llm-mock")
	req.ShortcutPath = filepath.Join(tmp, "llm-mock-run-opencode")

	buildMain := exec.Command("go", "build", "-o", req.BinaryPath, "./agent/llm/llm-mock")
	buildMain.Dir = req.RepoRoot
	if out, err := buildMain.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock: %w\n%s", err, string(out))
	}

	shortcutMain := filepath.Join(req.RepoRoot, "agent/llm/llm-mock/llm-mock-run-opencode/main.go")
	if _, err := os.Stat(shortcutMain); err == nil {
		buildShortcut := exec.Command("go", "build", "-o", req.ShortcutPath, "./agent/llm/llm-mock/llm-mock-run-opencode")
		buildShortcut.Dir = req.RepoRoot
		if out, err := buildShortcut.CombinedOutput(); err != nil {
			return fmt.Errorf("build llm-mock-run-opencode: %w\n%s", err, string(out))
		}
	} else {
		req.ShortcutPath = ""
	}

	if req.WorkDir == "" {
		req.WorkDir = filepath.Join(tmp, "work")
	}
	return nil
}

func minimalMockConfigJSON(port int, exchanges string) string {
	if port == 0 {
		port = 8080
	}
	if exchanges == "" {
		exchanges = `[
    {
      "request": {"content": "config-only-prompt", "index": -1},
      "response": {"content": "from-config", "finish_reason": "stop"}
    }
  ]`
	}
	return fmt.Sprintf(`{
  "port": %d,
  "exchanges": %s
}`, port, exchanges)
}

func envWithOverrides(base []string, overrides map[string]string) []string {
	merged := make(map[string]string, len(base)+len(overrides))
	for _, kv := range base {
		key, value, ok := strings.Cut(kv, "=")
		if !ok || key == "" {
			continue
		}
		merged[key] = value
	}
	for key, value := range overrides {
		merged[key] = value
	}
	out := make([]string, 0, len(merged))
	for key, value := range merged {
		out = append(out, key+"="+value)
	}
	return out
}

func parseOpencodeConfigDirFromOutput(combined string) string {
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "OPENCODE_CONFIG_DIR=") {
			return strings.TrimPrefix(line, "OPENCODE_CONFIG_DIR=")
		}
	}
	re := regexpOpencodeConfigDir()
	if m := re.FindStringSubmatch(combined); len(m) > 1 {
		return m[1]
	}
	return ""
}

func parseOpencodeHomeFromOutput(combined string) string {
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HOME=") {
			return strings.TrimPrefix(line, "HOME=")
		}
	}
	re := regexp.MustCompile(`HOME=([^\s]+)`)
	if m := re.FindStringSubmatch(combined); len(m) > 1 {
		return m[1]
	}
	return ""
}

func regexpOpencodeConfigDir() *regexp.Regexp {
	return regexp.MustCompile(`OPENCODE_CONFIG_DIR=([^\s]+)`)
}

func parseOpencodeConfigContentFromEnv(env []string) string {
	for _, kv := range env {
		key, value, ok := strings.Cut(kv, "=")
		if ok && key == "OPENCODE_CONFIG_CONTENT" {
			return value
		}
	}
	return ""
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil && resp.ExitCode == 0 {
		t.Fatalf("run failed: %v\nstderr: %s", resp.Err, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got string, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}

func readJSONLLinesFromContent(text string) ([]string, error) {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func parseAgentEventMaps(lines []string) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(lines))
	for i, line := range lines {
		var ev map[string]any
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return nil, fmt.Errorf("line %d: invalid JSONL: %w\n%s", i+1, err, line)
		}
		typ, _ := ev["type"].(string)
		if typ == "" {
			return nil, fmt.Errorf("line %d: missing or empty type field in %#v", i+1, ev)
		}
		if _, hasMethod := ev["method"]; hasMethod {
			return nil, fmt.Errorf("line %d: RecordedRequest shape (method) not AgentEvent: %#v", i+1, ev)
		}
		if _, hasPath := ev["path"]; hasPath {
			return nil, fmt.Errorf("line %d: RecordedRequest shape (path) not AgentEvent: %#v", i+1, ev)
		}
		out = append(out, ev)
	}
	return out, nil
}

func parseHTTPExchangeMaps(lines []string) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(lines))
	for i, line := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, fmt.Errorf("line %d: invalid JSONL: %w\n%s", i+1, err, line)
		}
		reqObj, ok := rec["request"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("line %d: missing request object in %#v", i+1, rec)
		}
		respObj, ok := rec["response"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("line %d: missing response object in %#v", i+1, rec)
		}
		method, _ := reqObj["method"].(string)
		if method == "" {
			return nil, fmt.Errorf("line %d: missing request.method in %#v", i+1, rec)
		}
		path, _ := reqObj["path"].(string)
		if path == "" {
			return nil, fmt.Errorf("line %d: missing request.path in %#v", i+1, rec)
		}
		if _, hasStatus := respObj["status"]; !hasStatus {
			return nil, fmt.Errorf("line %d: missing response.status in %#v", i+1, rec)
		}
		if typ, hasType := rec["type"]; hasType && typ != "" {
			return nil, fmt.Errorf("line %d: AgentEvent shape (type) not HTTP exchange: %#v", i+1, rec)
		}
		out = append(out, rec)
	}
	return out, nil
}

func agentEventsHaveTypes(events []map[string]any, want ...string) (bool, string) {
	found := make(map[string]bool, len(want))
	for _, ev := range events {
		typ, _ := ev["type"].(string)
		for _, w := range want {
			if typ == w {
				found[w] = true
			}
		}
	}
	var missing []string
	for _, w := range want {
		if !found[w] {
			missing = append(missing, w)
		}
	}
	if len(missing) == 0 {
		return true, ""
	}
	return false, strings.Join(missing, ", ")
}

func httpLogHasChatCompletionsModel(records []map[string]any, models ...string) bool {
	for _, rec := range records {
		reqObj, ok := rec["request"].(map[string]any)
		if !ok {
			continue
		}
		path, _ := reqObj["path"].(string)
		if !strings.Contains(path, "/v1/chat/completions") {
			continue
		}
		bodyRaw, ok := reqObj["body"]
		if !ok {
			continue
		}
		var body map[string]any
		switch v := bodyRaw.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &body); err != nil {
				continue
			}
		case map[string]any:
			body = v
		default:
			continue
		}
		model, _ := body["model"].(string)
		for _, want := range models {
			if model == want {
				return true
			}
		}
	}
	return false
}

func installFakeOpencodeEchoArgv(t *testing.T, req *Request) error {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "fake-opencode-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir fake opencode bin: %w", err)
	}
	script := "#!/bin/sh\nprintf 'OPENCODE_ARGV=%s\\n' \"$*\"\nprintf 'OPENCODE_CONFIG_DIR=%s\\n' \"$OPENCODE_CONFIG_DIR\"\nexit 0\n"
	opencodePath := filepath.Join(binDir, "opencode")
	if err := os.WriteFile(opencodePath, []byte(script), 0755); err != nil {
		return fmt.Errorf("write fake opencode: %w", err)
	}
	req.OpencodePathPrepend = binDir
	req.FakeOpencodeCmd = ""
	return nil
}
```