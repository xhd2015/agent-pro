# Scenario

**Feature**: `llm-mock run grok` orchestrator — mock server + isolated grok home + grok foreground

```
# resolve config + optional events merge
orchestrator -> mockconfig loader -> JSON config + events JSONL

# background mock, foreground grok
orchestrator -> llm-mock HTTP server (background)
orchestrator -> write GROK_HOME/config.toml (models_base_url -> mock)
orchestrator -> grok CLI (foreground, fake or real)

# teardown + session proof
grok exit <- orchestrator tears down mock
integration <- events.jsonl turn_started model_id mock-model
```

## Preconditions

- Repository contains `agent/llm/llm-mock` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Root `Setup` builds both binaries into `t.TempDir()`.
- Default plumbing tests set `LLM_MOCK_RUN_GROK_COMMAND` to a fake grok shell hook.
- `integration/` leaves require real `grok` on PATH (`label: real-grok, slow`).

## Steps

1. Root `Setup` resolves `RepoRoot`, builds `llm-mock` and `llm-mock-run-grok`.
2. Grouping `Setup` narrows config/events/home/CLI/integration profile.
3. Leaf `Setup` sets `Request` fields (config JSON, env mode, fake grok, grok args).
4. `Run` writes temp config/events files, sets env, executes orchestrator entrypoint.
5. Leaf `Assert` checks exit code, stdout/stderr, `config.toml`, or grok session events.

## Context

- `Request.ConfigEnv` — `"file"` (`LLM_MOCK_CONFIG_FILE`), `"legacy"` (`LLM_MOCK_CONFIG`), or `""` (neither).
- `Request.FakeGrokCmd` — replaces grok binary for plumbing tests.
- `Request.UseShortcut` — `true` runs `llm-mock-run-grok` instead of `llm-mock run grok`.
- `Response.GrokHomeUsed` — parsed from fake grok stdout or explicit `LLM_MOCK_GROK_HOME`.
- `FindNewestGrokSessionEvents` — locates newest `events.jsonl` under encoded workdir.

```go
import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const fakeGrokPrintHome = `sh -c 'echo GROK_HOME=$GROK_HOME; exit 0'`

const fakeGrokCurlOnce = `sh -c '
base="${GROK_MODELS_BASE_URL}"
body=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"config-only-prompt\"}]}")
echo "$body"
'`

const fakeGrokCurlTwice = `sh -c '
base="${GROK_MODELS_BASE_URL}"
r1=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"config-first-prompt\"}]}")
r2=$(curl -sf "$base/chat/completions" -H "Content-Type: application/json" -d "{\"model\":\"mock-model\",\"messages\":[{\"role\":\"user\",\"content\":\"events-second-prompt\"}]}")
echo "R1=$r1"
echo "R2=$r2"
'`

func Setup(t *testing.T, req *Request) error {
	_ = assertSuccess
	_ = assertContains
	_ = assertNotContains
	_ = assertExitCode
	_ = envWithOverrides
	_ = parseGrokHomeFromOutput
	_ = FindNewestGrokSessionEvents
	_ = grokEventsHasTurnStarted
	_ = minimalMockConfigJSON
	_ = fakeGrokPrintHome
	_ = fakeGrokCurlOnce
	_ = fakeGrokCurlTwice
	_ = readJSONLLinesFromContent
	_ = parseRecordedRequestMaps
	_ = parseAgentEventMaps
	_ = parseHTTPExchangeMaps
	_ = installFakeGrokEchoArgv

	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}

	tmp := t.TempDir()
	req.BinaryPath = filepath.Join(tmp, "llm-mock")
	req.ShortcutPath = filepath.Join(tmp, "llm-mock-run-grok")

	buildMain := exec.Command("go", "build", "-buildvcs=false", "-o", req.BinaryPath, "./agent/llm/llm-mock")
	buildMain.Dir = req.RepoRoot
	if out, err := buildMain.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock: %w\n%s", err, string(out))
	}

	buildShortcut := exec.Command("go", "build", "-buildvcs=false", "-o", req.ShortcutPath, "./agent/llm/llm-mock/llm-mock-run-grok")
	buildShortcut.Dir = req.RepoRoot
	if out, err := buildShortcut.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-grok: %w\n%s", err, string(out))
	}

	if req.WorkDir == "" {
		req.WorkDir = filepath.Join(tmp, "work")
	}
	if req.ExpectedExit == 0 && req.ConfigEnv != "" {
		// default success unless leaf overrides
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

func parseGrokHomeFromOutput(combined string) string {
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "GROK_HOME=") {
			return strings.TrimPrefix(line, "GROK_HOME=")
		}
	}
	re := regexpGrokHome()
	if m := re.FindStringSubmatch(combined); len(m) > 1 {
		return m[1]
	}
	return ""
}

func regexpGrokHome() *regexp.Regexp {
	return regexp.MustCompile(`GROK_HOME=([^\s]+)`)
}

// FindNewestGrokSessionEvents returns the newest session dir and events.jsonl lines
// under grokHome/sessions/<url-encoded-abs-workDir>/.
func FindNewestGrokSessionEvents(grokHome, workDir string) (sessionDir, eventsPath string, lines []string, err error) {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return "", "", nil, fmt.Errorf("abs workdir: %w", err)
	}
	encoded := url.PathEscape(abs)
	sessionsRoot := filepath.Join(grokHome, "sessions", encoded)
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return "", "", nil, fmt.Errorf("read sessions root %s: %w", sessionsRoot, err)
	}

	type candidate struct {
		dir  string
		mtime int64
	}
	var candidates []candidate
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		dir := filepath.Join(sessionsRoot, ent.Name())
		eventsFile := filepath.Join(dir, "events.jsonl")
		if st, statErr := os.Stat(eventsFile); statErr == nil {
			candidates = append(candidates, candidate{dir: dir, mtime: st.ModTime().UnixNano()})
			continue
		}
		if st, statErr := os.Stat(dir); statErr == nil {
			candidates = append(candidates, candidate{dir: dir, mtime: st.ModTime().UnixNano()})
		}
	}
	if len(candidates) == 0 {
		return "", "", nil, fmt.Errorf("no session dirs under %s", sessionsRoot)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].mtime > candidates[j].mtime
	})
	sessionDir = candidates[0].dir
	eventsPath = filepath.Join(sessionDir, "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		return sessionDir, eventsPath, nil, fmt.Errorf("read events.jsonl: %w", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return sessionDir, eventsPath, lines, scanner.Err()
}

func grokEventsHasEventType(lines []string, eventType string) bool {
	for _, line := range lines {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		if typ, _ := ev["type"].(string); typ == eventType {
			return true
		}
	}
	return false
}

func grokEventsHasTurnStarted(lines []string, modelID string) bool {
	for _, line := range lines {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		typ, _ := ev["type"].(string)
		if typ != "turn_started" {
			continue
		}
		if mid, _ := ev["model_id"].(string); mid == modelID {
			return true
		}
	}
	return false
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

// readJSONLLinesFromContent splits trimmed JSONL text into non-empty lines.
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

// parseRecordedRequestMaps parses JSONL lines into RecordedRequest-shaped objects.
func parseRecordedRequestMaps(lines []string) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, fmt.Errorf("invalid JSONL line: %w\n%s", err, line)
		}
		out = append(out, rec)
	}
	return out, nil
}

// parseAgentEventMaps parses JSONL lines into agent/event/types AgentEvent-shaped objects.
// Each line must have a non-empty string "type" field (think, message, tool_call, ...).
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

// parseHTTPExchangeMaps parses --log-http JSONL lines into HTTP exchange records.
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

// installFakeGrokImmediateExit prepends a fake grok on PATH that exits 0 (no session dirs).
func installFakeGrokImmediateExit(t *testing.T, req *Request) error {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "fake-grok-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir fake grok bin: %w", err)
	}
	script := "#!/bin/sh\nexit 0\n"
	grokPath := filepath.Join(binDir, "grok")
	if err := os.WriteFile(grokPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("write fake grok: %w", err)
	}
	req.GrokPathPrepend = binDir
	req.FakeGrokCmd = ""
	return nil
}

// installFakeGrokEmptySessionExit prepends a fake grok that creates a session dir without
// events.jsonl then exits 0 — reproduces interactive /exit leaving an empty session tree.
func installFakeGrokEmptySessionExit(t *testing.T, req *Request) error {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "fake-grok-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir fake grok bin: %w", err)
	}
	script := `#!/bin/sh
wd=$(pwd)
enc=$(python3 -c '
import os, sys, urllib.parse
p = os.path.abspath(sys.argv[1])
if p.startswith("/private/var/"):
    p = "/var/" + p[len("/private/var/"):]
elif p.startswith("/private/tmp/"):
    p = "/tmp/" + p[len("/private/tmp/"):]
print(urllib.parse.quote(p, safe=""))
' "$wd")
mkdir -p "$GROK_HOME/sessions/$enc/no-events-session"
exit 0
`
	grokPath := filepath.Join(binDir, "grok")
	if err := os.WriteFile(grokPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("write fake grok: %w", err)
	}
	req.GrokPathPrepend = binDir
	req.FakeGrokCmd = ""
	return nil
}

// installFakeGrokEchoArgv prepends a fake `grok` script to PATH that prints argv and GROK_HOME.
func installFakeGrokEchoArgv(t *testing.T, req *Request) error {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "fake-grok-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir fake grok bin: %w", err)
	}
	script := "#!/bin/sh\nprintf 'GROK_ARGV=%s\\n' \"$*\"\nprintf 'GROK_HOME=%s\\n' \"$GROK_HOME\"\nexit 0\n"
	grokPath := filepath.Join(binDir, "grok")
	if err := os.WriteFile(grokPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("write fake grok: %w", err)
	}
	req.GrokPathPrepend = binDir
	req.FakeGrokCmd = ""
	return nil
}
```
