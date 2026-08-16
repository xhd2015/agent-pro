# Scenario

**Feature**: `llm-mock run codex` orchestrator — mock server + isolated codex home + codex foreground

```
# resolve config
orchestrator -> mockconfig loader -> JSON config

# background mock, foreground codex
orchestrator -> llm-mock HTTP server (background)
orchestrator -> write CODEX_HOME/config.toml (base_url -> mock, wire_api=responses)
orchestrator -> codex CLI (foreground, fake or real)

# teardown
codex exit <- orchestrator tears down mock (no session mirror poll)
integration <- stdout Paris from mocked Responses API
```

## Preconditions

- Repository contains `agent/llm/llm-mock` and `agent/llm/llm-mock/llm-mock-run-codex`.
- Root `Setup` builds both binaries into `t.TempDir()`.
- Default plumbing tests set `LLM_MOCK_RUN_CODEX_COMMAND` to a fake codex shell hook.
- `integration/` leaves require real `codex` on PATH (`label: real-codex, slow`).

## Steps

1. Root `Setup` resolves `RepoRoot`, builds `llm-mock` and `llm-mock-run-codex`.
2. Grouping `Setup` narrows config/home/CLI/integration profile.
3. Leaf `Setup` sets `Request` fields (config JSON, env mode, fake codex, codex args).
4. `Run` writes temp config files, sets env, executes orchestrator entrypoint.
5. Leaf `Assert` checks exit code, stdout/stderr, `config.toml`, or log JSONL files.

## Context

- `Request.ConfigEnv` — `"file"` (`LLM_MOCK_CONFIG_FILE`) or `""` (neither).
- `Request.FakeCodexCmd` — replaces codex binary for plumbing tests.
- `Request.UseShortcut` — `true` runs `llm-mock-run-codex` instead of `llm-mock run codex`.
- `Response.CodexHomeUsed` — parsed from fake codex stdout/stderr or explicit `LLM_MOCK_CODEX_HOME`.
- Fake codex curls read `base_url` from `$CODEX_HOME/config.toml` (no `GROK_MODELS_BASE_URL` equivalent).

```go
import (
	"runtime"
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

const fakeCodexPrintHome = `sh -c 'echo CODEX_HOME=$CODEX_HOME; exit 0'`

const fakeCodexCurlResponsesOnce = `sh -c '
base=$(grep -m1 base_url "$CODEX_HOME/config.toml" | sed -n "s/.*= *\"\([^\"]*\)\".*/\1/p")
body=$(curl -sf "$base/responses" -H "Content-Type: application/json" -H "Authorization: Bearer $OPENAI_API_KEY" -d "{\"model\":\"mock-model\",\"input\":[{\"role\":\"user\",\"content\":\"config-only-prompt\"}]}")
echo "$body"
'`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = assertSuccess
	_ = assertContains
	_ = assertNotContains
	_ = assertExitCode
	_ = envWithOverrides
	_ = parseCodexHomeFromOutput
	_ = minimalMockConfigJSON
	_ = fakeCodexPrintHome
	_ = fakeCodexCurlResponsesOnce
	_ = readJSONLLinesFromContent
	_ = parseAgentEventMaps
	_ = parseHTTPExchangeMaps
	_ = installFakeCodexEchoArgv
	_ = agentEventsHaveTypes

	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}

	tmp := t.TempDir()
	req.BinaryPath = filepath.Join(tmp, "llm-mock")
	req.ShortcutPath = filepath.Join(tmp, "llm-mock-run-codex")

	buildMain := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.BinaryPath, "./agent/llm/llm-mock")
	buildMain.Dir = req.RepoRoot
	if out, err := buildMain.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock: %w\n%s", err, string(out))
	}

	buildShortcut := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.ShortcutPath, "./agent/llm/llm-mock/llm-mock-run-codex")
	buildShortcut.Dir = req.RepoRoot
	if out, err := buildShortcut.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-codex: %w\n%s", err, string(out))
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
      "request": {"role": "user", "content": "config-only-prompt", "index": -1},
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

func parseCodexHomeFromOutput(combined string) string {
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "CODEX_HOME=") {
			return strings.TrimPrefix(line, "CODEX_HOME=")
		}
	}
	re := regexpCodexHome()
	if m := re.FindStringSubmatch(combined); len(m) > 1 {
		return m[1]
	}
	return ""
}

func regexpCodexHome() *regexp.Regexp {
	return regexp.MustCompile(`CODEX_HOME=([^\s]+)`)
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

func installFakeCodexEchoArgv(t *testing.T, req *Request) error {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "fake-codex-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir fake codex bin: %w", err)
	}
	script := "#!/bin/sh\nprintf 'CODEX_ARGV=%s\\n' \"$*\"\nprintf 'CODEX_HOME=%s\\n' \"$CODEX_HOME\"\nexit 0\n"
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("write fake codex: %w", err)
	}
	req.CodexPathPrepend = binDir
	req.FakeCodexCmd = ""
	return nil
}
```