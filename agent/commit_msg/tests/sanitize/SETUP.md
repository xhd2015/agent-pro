# Scenario

**Feature**: post-parse sanitize strips commit-message anti-patterns before stdout and `--commit`

```
# agent text may be dirty (outer ticks, md meta, git -m wrappers, tool noise)
staged diff -> gen-commit-msg -> fake-opencode dirty payload
  -> parseCommitMsgFromText -> sanitize (NEW) -> format -> stdout
  -> optional --commit uses sanitized message only

# unusable garbage after sanitize hard-fails (no commit, non-zero)
```

## Preconditions
- Shared anti-pattern fixtures live at `agent/commit_msg/testdata/anti_patterns/`.
- Leaves feed fixture `.in` text through fake-opencode as the agent message payload.
- Sanitize is not implemented yet (classic TDD: leaves assert desired clean / reject behavior → RED).

## Steps
1. Resolve the anti-patterns fixture directory relative to `DOCTEST_ROOT`.
2. Provide helpers to load fixtures and write escaped mock agent text.
3. Leaves stage a git repo and set `req.Commit` when the leaf needs git subject / HEAD checks.

## Context
- Fixture naming: `<name>.in` (raw agent text), `<name>.want` (formatted message), or `<name>.want_err` (reject substring).
- `WriteMockAgentText` JSON-encodes the agent payload so backticks and quotes survive mock config.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

// antiPatternsDirAbs is set once per Setup from d.DOCTEST_ROOT (fixture path is immutable).
var antiPatternsDirAbs string

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = AntiPatternsDir
	_ = ReadAntiPatternIn
	_ = ReadAntiPatternWant
	_ = ReadAntiPatternWantErr
	_ = HasAntiPatternWantErr
	_ = WriteMockAgentText
	_ = StageRepoWithChange
	_ = AssertStdoutMessage
	_ = AssertGitSubject
	_ = AssertHEADUnchanged
	_ = ListAntiPatternNames
	_ = RunAntiPatternFixture
	if req.TempDir == "" {
		return fmt.Errorf("sanitize subtree requires initialized TempDir from root Setup")
	}
	// agent/commit_msg/testdata/anti_patterns (sibling of tests/)
	antiPatternsDirAbs = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "testdata", "anti_patterns"))
	return nil
}

// AntiPatternsDir is agent/commit_msg/testdata/anti_patterns (sibling of tests/).
func AntiPatternsDir() string {
	return antiPatternsDirAbs
}

func ReadAntiPatternIn(t *testing.T, name string) string {
	t.Helper()
	return readAntiPatternFile(t, name+".in")
}

func ReadAntiPatternWant(t *testing.T, name string) string {
	t.Helper()
	return strings.TrimRight(readAntiPatternFile(t, name+".want"), "\n")
}

func ReadAntiPatternWantErr(t *testing.T, name string) string {
	t.Helper()
	return strings.TrimSpace(readAntiPatternFile(t, name+".want_err"))
}

func HasAntiPatternWantErr(name string) bool {
	_, err := os.Stat(filepath.Join(AntiPatternsDir(), name+".want_err"))
	return err == nil
}

func readAntiPatternFile(t *testing.T, file string) string {
	t.Helper()
	path := filepath.Join(AntiPatternsDir(), file)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(b)
}

// WriteMockAgentText builds a fake-opencode mock that emits agentText as the step message.
func WriteMockAgentText(t *testing.T, req *Request, sessionID, agentText string) {
	t.Helper()
	type evt struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	}
	cfg := map[string]interface{}{
		"version":    "agent-pro.fake-runner.v1",
		"runner":     "fake-opencode",
		"session_id": sessionID,
		"llm_events": []evt{
			{Type: "step_start"},
			{Type: "message", Text: agentText},
			{Type: "step_finish"},
		},
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal mock config: %v", err)
	}
	WriteMockConfig(t, req, string(b))
}

func StageRepoWithChange(t *testing.T, req *Request) {
	t.Helper()
	if req.GitDir == "" {
		req.GitDir = filepath.Join(req.TempDir, "repo")
	}
	InitGitRepo(t, req.GitDir)
	WriteFile(t, filepath.Join(req.GitDir, "sanitize_probe.go"), "package main\n// probe\n")
	RunGit(t, req.GitDir, "add", "sanitize_probe.go")
}

// ListAntiPatternNames returns basenames that have a .in file.
func ListAntiPatternNames(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(AntiPatternsDir())
	if err != nil {
		t.Fatalf("read anti_patterns dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".in") {
			names = append(names, strings.TrimSuffix(name, ".in"))
		}
	}
	if len(names) == 0 {
		t.Fatal("no anti_patterns fixtures found")
	}
	return names
}

// RunAntiPatternFixture stages (if needed), loads fixture name, runs gen-commit-msg once.
func RunAntiPatternFixture(t *testing.T, req *Request, name string) *Response {
	t.Helper()
	if req.GitDir == "" {
		StageRepoWithChange(t, req)
	}
	in := ReadAntiPatternIn(t, name)
	WriteMockAgentText(t, req, "sess_sanitize_"+name, in)
	resp, err := captureRunGenCommitMsg(t, req)
	if err != nil {
		t.Fatalf("captureRunGenCommitMsg for %s: %v", name, err)
	}
	return resp
}

// AssertStdoutMessage checks stdout equals formatted message plus trailing newline.
func AssertStdoutMessage(t *testing.T, stdout, want string) {
	t.Helper()
	wantOut := want + "\n"
	if stdout != wantOut {
		t.Fatalf("stdout mismatch\n got: %q\nwant: %q", stdout, wantOut)
	}
}

func AssertGitSubject(t *testing.T, gitDir, wantSubject string) {
	t.Helper()
	out, err := execGitOutput(gitDir, "log", "-1", "--format=%s")
	if err != nil {
		t.Fatalf("git log -1 --format=%%s: %v", err)
	}
	got := strings.TrimSpace(out)
	if got != wantSubject {
		t.Fatalf("git subject = %q, want %q", got, wantSubject)
	}
}

func AssertHEADUnchanged(t *testing.T, gitDir, headBefore string) {
	t.Helper()
	headAfter, err := execGitOutput(gitDir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	if strings.TrimSpace(headAfter) != strings.TrimSpace(headBefore) {
		t.Fatalf("HEAD changed on rejected sanitize: before=%s after=%s",
			strings.TrimSpace(headBefore), strings.TrimSpace(headAfter))
	}
}

func execGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}
```
