package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// --- test framework ---

type T struct {
	name   string
	failed bool
	logs   []string
}

func (t *T) Logf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	t.logs = append(t.logs, s)
}

func (t *T) Errorf(format string, args ...interface{}) {
	t.failed = true
	msg := fmt.Sprintf(format, args...)
	t.logs = append(t.logs, "ERROR: "+msg)
}

func (t *T) Fatalf(format string, args ...interface{}) {
	t.Errorf(format, args...)
	panic(testFail{t})
}

func (t *T) Helper() {}

type testFail struct{ t *T }

type testCase struct {
	name string
	fn   func(t *T)
}

var (
	explainBin  string
	tmpBase     string
	fakeBinDir  string
	sessionHome string
	passed      int
	failed      int
	failDetails []string
)

func main() {
	var err error
	tmpBase, err = os.MkdirTemp("", "explain-integ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create tmp base: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpBase)

	sessionHome = filepath.Join(tmpBase, "config-home")
	fakeBinDir = filepath.Join(tmpBase, "bin")
	if err := os.MkdirAll(fakeBinDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "create fake bin dir: %v\n", err)
		os.Exit(1)
	}

	if err := writeFakeOpencode(fakeBinDir); err != nil {
		fmt.Fprintf(os.Stderr, "write fake opencode: %v\n", err)
		os.Exit(1)
	}

	explainBin, err = buildExplain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build explain: %v\n", err)
		os.Exit(1)
	}

	tests := []testCase{
		{"TestSingleArgNewSession", testSingleArgNewSession},
		{"TestSingleArgRepeatNewSession", testSingleArgRepeatNewSession},
		{"TestTwoArgsExactMatchResume", testTwoArgsExactMatchResume},
		{"TestTwoArgsNoMatchNewSession", testTwoArgsNoMatchNewSession},
		{"TestThreeArgsTwoPrefixMatch", testThreeArgsTwoPrefixMatch},
		{"TestVerboseFlag", testVerboseFlag},
		{"TestModelFlagStored", testModelFlagStored},
		{"TestFirstEverRun", testFirstEverRun},
		{"TestChineseAndEmojiPrompt", testChineseAndEmojiPrompt},
	}

	for _, tc := range tests {
		runTest(tc)
	}

	fmt.Println()
	if failed > 0 {
		fmt.Printf("%d passed, %d failed\n", passed, failed)
		for _, d := range failDetails {
			fmt.Println(d)
		}
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Printf("%d passed\n", passed)
	fmt.Println("PASS")
}

func runTest(tc testCase) {
	t := &T{name: tc.name}
	fmt.Printf("=== RUN   %s\n", tc.name)

	defer func() {
		r := recover()
		if r != nil {
			if sf, ok := r.(testFail); ok {
				for _, l := range sf.t.logs {
					fmt.Printf("    %s\n", l)
				}
				fmt.Printf("--- FAIL: %s\n", tc.name)
				failed++
				failDetails = append(failDetails, fmt.Sprintf("  FAIL: %s", tc.name))
			} else {
				panic(r)
			}
			return
		}
		if t.failed {
			for _, l := range t.logs {
				fmt.Printf("    %s\n", l)
			}
			fmt.Printf("--- FAIL: %s\n", tc.name)
			failed++
			failDetails = append(failDetails, fmt.Sprintf("  FAIL: %s", tc.name))
		} else {
			fmt.Printf("--- PASS: %s\n", tc.name)
			passed++
		}
	}()

	resetState()
	tc.fn(t)
}

func resetState() {
	sessDir := filepath.Join(sessionHome, "sessions")
	os.RemoveAll(sessDir)
	resetFakeOpencode()
}

// --- helpers ---

func buildExplain() (string, error) {
	modRoot := findModuleRoot()
	if modRoot == "" {
		return "", fmt.Errorf("cannot find module root")
	}
	bin := filepath.Join(tmpBase, "explain")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/explain")
	cmd.Dir = modRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return bin, nil
}

func findModuleRoot() string {
	d, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		d = parent
	}
}

func writeFakeOpencode(binDir string) error {
	script := fmt.Sprintf(`#!/bin/bash
# Fake opencode for integration tests
ARG1="$1"
shift

case "$ARG1" in
  run)
    CNT=$(cat "%s" 2>/dev/null || echo 0)
    CNT=$((CNT + 1))
    mkdir -p "%s"
    echo "$CNT" > "%s"

    SID="fake-sess-${CNT}"
    echo '{"type":"step_start","sessionID":"'"$SID"'","timestamp":0}'
    echo '{"type":"text","part":{"text":"[MOCK '"$CNT"']"}}'
    echo '{"type":"step_finish","sessionID":"'"$SID"'","timestamp":0}'
    ;;
  models)
    echo "deepseek-v4-pro"
    ;;
  *)
    exit 1
    ;;
esac
`, getCounterPath(), fakeBinDir, getCounterPath())

	path := filepath.Join(binDir, "opencode")
	return os.WriteFile(path, []byte(script), 0755)
}

func resetFakeOpencode() {
	os.Remove(getCounterPath())
}

func getCounterPath() string {
	return filepath.Join(tmpBase, "fake-opencode-counter")
}

func explainEnv() []string {
	env := os.Environ()
	env = append(env,
		"AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME="+sessionHome,
		"EXPLAIN_AGENT_PATH="+filepath.Join(fakeBinDir, "opencode"),
	)
	return env
}

func runExplain(t *T, args ...string) (string, string) {
	t.Helper()
	cmd := exec.Command(explainBin, args...)
	cmd.Env = explainEnv()
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	errStr := strings.TrimSpace(stderr.String())
	t.Logf("stdout: %s, stderr: %s", out, errStr)
	if err != nil {
		t.Fatalf("explain failed: %v\n%s", err, errStr)
	}
	return out, errStr
}

func listSessionDirs(t *T) []string {
	t.Helper()
	sessDir := filepath.Join(sessionHome, "sessions")
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read sessions dir: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

func readSessionData(t *T, dirName string) SessionData {
	t.Helper()
	path := filepath.Join(sessionHome, "sessions", dirName, "session.data")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read session.data: %v", err)
	}
	var data SessionData
	if err := json.Unmarshal(bytes, &data); err != nil {
		t.Fatalf("unmarshal session.data: %v", err)
	}
	return data
}

func assertEqual(t *T, name string, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", name, expected, actual)
	}
}

func assertTrue(t *T, name string, cond bool) {
	t.Helper()
	if !cond {
		t.Errorf("%s: expected true", name)
	}
}

// --- types ---

type SessionData struct {
	AgentRunner string     `json:"agent_runner"`
	Model       string     `json:"model"`
	Messages    []Message  `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

// --- tests ---

func testSingleArgNewSession(t *T) {
	out, _ := runExplain(t, "hello")
	assertTrue(t, "output contains MOCK", strings.Contains(out, "[MOCK 1]"))

	dirs := listSessionDirs(t)
	assertEqual(t, "session count", 1, len(dirs))

	data := readSessionData(t, dirs[0])
	assertEqual(t, "agent runner", "opencode", data.AgentRunner)
	assertEqual(t, "message count", 2, len(data.Messages))
	assertEqual(t, "msg[0] role", "user", data.Messages[0].Role)
	assertEqual(t, "msg[0] content", "hello", data.Messages[0].Message)
	assertEqual(t, "msg[1] role", "assistant", data.Messages[1].Role)
}

func testSingleArgRepeatNewSession(t *T) {
	runExplain(t, "hello")
	runExplain(t, "hello")

	dirs := listSessionDirs(t)
	assertEqual(t, "session count", 2, len(dirs))
}

func testTwoArgsExactMatchResume(t *T) {
	runExplain(t, "A F", "B")
	dirs1 := listSessionDirs(t)
	assertEqual(t, "first call dirs", 1, len(dirs1))
	data1 := readSessionData(t, dirs1[0])
	assertEqual(t, "first msgs count", 4, len(data1.Messages))

	runExplain(t, "A F", "C")

	dirs2 := listSessionDirs(t)
	assertEqual(t, "still one dir", 1, len(dirs2))

	data2 := readSessionData(t, dirs2[0])
	assertEqual(t, "msg count after resume", 6, len(data2.Messages))
	assertEqual(t, "msg[4] (new user)", "user", data2.Messages[4].Role)
	assertEqual(t, "msg[4] content", "C", data2.Messages[4].Message)
	assertEqual(t, "msg[5] (new assistant)", "assistant", data2.Messages[5].Role)
}

func testTwoArgsNoMatchNewSession(t *T) {
	runExplain(t, "X", "Y")
	runExplain(t, "A", "B")

	dirs := listSessionDirs(t)
	assertEqual(t, "session count", 2, len(dirs))
}

func testThreeArgsTwoPrefixMatch(t *T) {
	runExplain(t, "A", "B")
	dirs1 := listSessionDirs(t)
	assertEqual(t, "first call dirs", 1, len(dirs1))
	data1 := readSessionData(t, dirs1[0])
	assertEqual(t, "first msgs count", 4, len(data1.Messages))

	runExplain(t, "A", "B", "C")

	dirs2 := listSessionDirs(t)
	assertEqual(t, "still one dir", 1, len(dirs2))

	data2 := readSessionData(t, dirs2[0])
	assertEqual(t, "msg count", 6, len(data2.Messages))
	assertEqual(t, "msg[4] content", "C", data2.Messages[4].Message)
}

func testVerboseFlag(t *T) {
	_, stderr := runExplain(t, "-v", "hello")
	assertTrue(t, "stderr contains [explain]", strings.Contains(stderr, "[explain]"))
}

func testModelFlagStored(t *T) {
	runExplain(t, "--model", "gpt-4", "hello")

	dirs := listSessionDirs(t)
	assertEqual(t, "session count", 1, len(dirs))

	data := readSessionData(t, dirs[0])
	assertEqual(t, "model", "gpt-4", data.Model)
}

func testFirstEverRun(t *T) {
	dirs := listSessionDirs(t)
	assertEqual(t, "no dirs before", 0, len(dirs))

	runExplain(t, "first ever")

	dirs = listSessionDirs(t)
	assertEqual(t, "one dir after", 1, len(dirs))

	data := readSessionData(t, dirs[0])
	assertEqual(t, "msg[0] role", "user", data.Messages[0].Role)
	assertEqual(t, "msg[0] content", "first ever", data.Messages[0].Message)
}

func testChineseAndEmojiPrompt(t *T) {
	runExplain(t, "run的各种形态", "nyaa")

	dirs := listSessionDirs(t)
	assertEqual(t, "session count", 1, len(dirs))

	data := readSessionData(t, dirs[0])
	assertEqual(t, "msgs", 4, len(data.Messages))
	assertEqual(t, "msg[0]", "run的各种形态", data.Messages[0].Message)
	assertEqual(t, "msg[2]", "nyaa", data.Messages[2].Message)

	assertTrue(t, "dir has hash", len(dirs[0]) > 28)
}
