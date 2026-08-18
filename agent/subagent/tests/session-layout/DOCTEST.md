# SessionLayout Tests

Verify `subagent.SessionLayout` controls where session files are written, which
artifacts are created, and how foreign `meta.json` schemas are merged.

## Version
0.0.2

# DSN (Domain Specific Notion)

The **subagent library** exposes `Run(ctx, Config, Options)` for agent invocation
and session file management. **SessionLayout** on `Options` overrides the session
root directory and per-file paths; feature flags disable questions FIFO and
progress reporting. When **Dir** is set, subagent skips date-nested `sess_*`
auto-creation and uses the caller-provided flat directory. **meta.json** merges agent fields (`opencode_session_id`, `explicit_session_id`
aliases) into pre-existing host-owned metadata when **`HostOwnsMeta`** is false
(legacy). When **`HostOwnsMeta`** is true, subagent skips meta reads/writes;
session match/resume trusts `SessionID` + `ResumeInnerSessionID`, and
**`OnAgentComplete(AgentRunInfo)`** reports the inner runner session id for the
host to persist.
**fake-codex** replays mock events; the **test harness** builds binaries, prepares
session dirs, invokes `subagent.Run`, and inspects filesystem artifacts.

## Decision Tree

```
session-layout/
├── DOCTEST.md
├── SETUP.md                                    # build fake-codex, shared mock + helpers
├── legacy/                                     # zero-value SessionLayout (Dir unset)
│   ├── SETUP.md                                # HOME-based nested sess_* layout
│   └── default-layout/
│       └── creates-nested-session/             # date/sess_* dir, questions, progress, messages
└── flat-dir/                                   # SessionLayout.Dir set (flat custom root)
    ├── SETUP.md                                # MkdirAll session dir, default layout flags
    ├── all-paths/                              # default paths + flags on
    ├── skip-messages/                          # MessagesPath="" → no messages.jsonl
    ├── disable-questions/                      # QuestionsEnabled=false
    ├── disable-progress/                       # ProgressEnabled=false
    ├── merged-meta/                            # foreign meta + opencode_session_id
    ├── default-flags-off/
    │   └── no-questions-no-progress/           # Dir set, zero flags → no questions/progress
    ├── custom-paths/
    │   └── events-outside-dir/                 # EventsPath override outside Dir
    ├── resume/
    │   └── second-run-appends-events/          # same SessionID appends events.jsonl
    ├── trace/
    │   └── custom-events-path/                 # traceSession reads custom EventsPath
    ├── meta-alias/
    │   └── agent-session-id-resolves/          # agent_session_id resolves flat session
    └── host-owns-meta/                         # HostOwnsMeta=true (task-hub flat layout)
        ├── no-meta-write/                      # subagent does not write meta.json
        ├── callback-receives-inner-id/         # OnAgentComplete(InnerSessionID)
        ├── resume-uses-resume-inner-id/        # ResumeInnerSessionID resume + append events
        ├── preserves-foreign-meta/             # merged-meta scenario; meta frozen
        └── preserves-meta-alias/               # meta-alias scenario; no explicit_session_id
```

Parameter ranking (most → least significant):
1. **Layout mode** — flat `Dir` vs legacy (zero value)
2. **Meta ownership** — `HostOwnsMeta` true (callback, no meta I/O) vs false (subagent merge)
3. **Feature flags** — `QuestionsEnabled`, `ProgressEnabled` (legacy defaults on; flat defaults off)
4. **Per-file overrides** — `EventsPath`, `MetaPath`, `MessagesPath`, etc.
5. **Session lifecycle** — create vs resume (`ResumeInnerSessionID`); trace reads same layout
6. **Meta merge** — pre-existing foreign schema and `agent_session_id` alias (HostOwnsMeta=false)

## Test Index

| Leaf | Branch | Description |
|------|--------|-------------|
| `legacy/default-layout/creates-nested-session` | legacy | Zero layout creates date-nested `sess_*` under HOME with questions, progress, messages |
| `flat-dir/all-paths` | flat-dir | Default layout writes events, messages, questions/, progress/ to custom dir |
| `flat-dir/skip-messages` | flat-dir | `MessagesPath=""` skips messages.jsonl |
| `flat-dir/disable-questions` | flat-dir | `QuestionsEnabled=false` skips questions/ and QUESTIONS stdout footer |
| `flat-dir/disable-progress` | flat-dir | `ProgressEnabled=false` skips progress/ |
| `flat-dir/merged-meta` | flat-dir | Host meta fields preserved; `opencode_session_id` updated after run |
| `flat-dir/default-flags-off/no-questions-no-progress` | flat-dir | Dir set with zero flags skips questions/ and progress/ |
| `flat-dir/custom-paths/events-outside-dir` | flat-dir | Custom `EventsPath` writes events outside `Dir` |
| `flat-dir/resume/second-run-appends-events` | flat-dir | Second run appends events and preserves host meta |
| `flat-dir/trace/custom-events-path` | flat-dir | `traceSession` reads custom `EventsPath` after run |
| `flat-dir/meta-alias/agent-session-id-resolves` | flat-dir | `agent_session_id` alias resolves session without `explicit_session_id` |
| `flat-dir/host-owns-meta/no-meta-write` | host-owns-meta | Pre-written meta bytes unchanged; subagent skips meta.json writes |
| `flat-dir/host-owns-meta/callback-receives-inner-id` | host-owns-meta | `OnAgentComplete` receives mock `InnerSessionID` and `AgentRunner` |
| `flat-dir/host-owns-meta/resume-uses-resume-inner-id` | host-owns-meta | Second run uses `ResumeInnerSessionID`; events append; meta frozen |
| `flat-dir/host-owns-meta/preserves-foreign-meta` | host-owns-meta | Foreign host meta preserved (merged-meta scenario, no subagent patch) |
| `flat-dir/host-owns-meta/preserves-meta-alias` | host-owns-meta | `agent_session_id` alias resume; no `explicit_session_id` or meta patch |

## How to Run

```sh
cd external/agent-pro-task-hub
doctest vet ./agent/subagent/tests/session-layout/
doctest test -v ./agent/subagent/tests/session-layout/
```

```go
import (
	"runtime"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	TempDir        string
	HomeDir        string
	FakeCodexBin   string
	SessionDir     string
	SessionID      string
	MockConfigPath string
	SecondMockConfigPath string
	PreCreateMeta  string
	CustomEventsPath string

	Layout subagent.SessionLayout

	HostOwnsMeta         bool
	ResumeInnerSessionID string
	MetaBytesBeforeRun   []byte
	OnAgentComplete      func(subagent.AgentRunInfo) error

	CallbackCalled         bool
	CallbackInnerSessionID string
	CallbackAgentRunner    string

	AgentRunner string
	Prompt      string
}

type Response struct {
	Stdout     string
	Stderr     string
	Err        error
	SessionDir string
}

func findLegacySessionDir(homeDir, roleName, sessionID string) (string, error) {
	base := filepath.Join(homeDir, ".agent-pro", "subagent", roleName, "sessions")
	var matches []string
	_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || !info.IsDir() || !strings.HasPrefix(info.Name(), "sess_") {
			return nil
		}
		metaPath := filepath.Join(path, "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			return nil
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return nil
		}
		if explicit, _ := m["explicit_session_id"].(string); explicit == sessionID {
			matches = append(matches, path)
		}
		return nil
	})
	if len(matches) == 0 {
		return "", fmt.Errorf("legacy session not found for %q under %s", sessionID, base)
	}
	return matches[len(matches)-1], nil
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.FakeCodexBin == "" {
		return nil, fmt.Errorf("FakeCodexBin not built — root Setup must run first")
	}
	if req.Layout.Dir != "" && req.SessionDir == "" {
		return nil, fmt.Errorf("SessionDir not set — flat-dir Setup must run first")
	}
	return invokeRun(t, req)
}

// residualChildEnvMu serializes PATH / AGENT_RUNNER_FAKE_CODEX_PATH process
// mutation required for registry path resolution (no product option path yet).
var residualChildEnvMu sync.Mutex

func invokeRun(t *testing.T, req *Request) (*Response, error) {
	t.Helper()

	// Residual process Setenv (serialized): registry resolves fake-codex via
	// os.Getenv / PATH LookPath. HOME uses Config.HomeDir (no process mutation).
	residualChildEnvMu.Lock()
	defer residualChildEnvMu.Unlock()

	binDir := filepath.Dir(req.FakeCodexBin)
	oldPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", binDir+string(filepath.ListSeparator)+oldPath)
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	prevFake, hadFake := os.LookupEnv("AGENT_RUNNER_FAKE_CODEX_PATH")
	if req.FakeCodexBin != "" {
		_ = os.Setenv("AGENT_RUNNER_FAKE_CODEX_PATH", req.FakeCodexBin)
		defer func() {
			if hadFake {
				_ = os.Setenv("AGENT_RUNNER_FAKE_CODEX_PATH", prevFake)
			} else {
				_ = os.Unsetenv("AGENT_RUNNER_FAKE_CODEX_PATH")
			}
		}()
	}

	var stdout, stderr bytes.Buffer
	opts := subagent.Options{
		Prompt:               req.Prompt,
		AgentRunner:          req.AgentRunner,
		MockConfig:           req.MockConfigPath,
		SessionID:            req.SessionID,
		SessionLayout:        req.Layout,
		StdoutWriter:         &stdout,
		HostOwnsMeta:         req.HostOwnsMeta,
		ResumeInnerSessionID: req.ResumeInnerSessionID,
	}
	if req.OnAgentComplete != nil {
		opts.OnAgentComplete = req.OnAgentComplete
	} else if req.HostOwnsMeta {
		opts.OnAgentComplete = func(info subagent.AgentRunInfo) error {
			req.CallbackCalled = true
			req.CallbackInnerSessionID = info.InnerSessionID
			req.CallbackAgentRunner = info.AgentRunner
			return nil
		}
	}

	runErr := subagent.Run(context.Background(), subagent.Config{
		RoleName:              "layout-test",
		AutoGenerateSessionID: false,
		PromptContent:         "You are a layout test agent.",
		HomeDir:               req.HomeDir,
	}, opts)

	sessionDir := req.SessionDir
	if sessionDir == "" && req.Layout.Dir == "" {
		found, err := findLegacySessionDir(req.HomeDir, "layout-test", req.SessionID)
		if err != nil {
			return nil, err
		}
		sessionDir = found
		req.SessionDir = found
	}

	return &Response{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Err:        runErr,
		SessionDir: sessionDir,
	}, runErr
}

func buildFakeCodex(t *testing.T, moduleRoot, out string) {
	t.Helper()
	cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", out, "./fake-codex")
	cmd.Dir = moduleRoot
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake-codex: %v\n%s", err, string(outBytes))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func minimalMockConfig(sessionID string) string {
	return fmt.Sprintf(`{
  "version": "agent-pro.fake-runner.v1",
  "runner": "fake-codex",
  "session_id": %q,
  "llm_events": [
    {"type": "message", "text": "layout test ok\n"}
  ]
}`, sessionID)
}

func resolvePath(layout subagent.SessionLayout, defName string, override string) string {
	if override != "" {
		return override
	}
	if layout.Dir != "" {
		return filepath.Join(layout.Dir, defName)
	}
	return defName
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func readMetaField(t *testing.T, metaPath, key string) string {
	t.Helper()
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta %s: %v", metaPath, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse meta: %v", err)
	}
	s, _ := m[key].(string)
	return s
}

func containsQuestionsFooter(s string) bool {
	return strings.Contains(s, "QUESTIONS")
}

func countJSONLLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	n := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

func assertNoNestedSessUnder(t *testing.T, root, exceptDir string) {
	t.Helper()
	exceptDir = filepath.Clean(exceptDir)
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || !info.IsDir() || !strings.HasPrefix(info.Name(), "sess_") {
			return nil
		}
		if filepath.Clean(path) != exceptDir && !strings.HasPrefix(filepath.Clean(path), exceptDir+string(filepath.Separator)) {
			t.Fatalf("unexpected nested sess_* dir outside %s: %s", exceptDir, path)
		}
		return nil
	})
}

func runTrace(t *testing.T, req *Request) (string, error) {
	t.Helper()

	oldOut := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := subagent.TestExported_traceSession(subagent.Config{
		RoleName: "layout-test",
	}, subagent.Options{
		CatchUp:       true,
		SessionID:     req.SessionID,
		SessionLayout: req.Layout,
	})

	wOut.Close()
	os.Stdout = oldOut

	var buf bytes.Buffer
	buf.ReadFrom(rOut)
	return buf.String(), err
}
```