// grok-tty-discovery-cancel reproduces the web grok-tty bind race locally.
//
// Uses the same harness as enhance-chat binding-failure: llm-mock-run-grok +
// real-like PTY chrome + empty GROK_HOME + --keep-tty (web path).
//
// Run from repo root:
//
//	go run ./script/debug/grok-tty-discovery-cancel
//	go run ./script/debug/grok-tty-discovery-cancel -scenario=delayed-session
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

const (
	prompt             = "one word of France capital"
	resolveThinkText   = "Resolve session id..."
	resolveErrorPrefix = "Cannot resolve session id:"
)

func main() {
	os.Exit(run())
}

func run() int {
	scenario := "chrome-cancel"
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-scenario=") {
			scenario = strings.TrimPrefix(arg, "-scenario=")
		}
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "repo root: %v\n", err)
		return 1
	}

	tmp, err := os.MkdirTemp("", "grok-discovery-cancel-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir temp: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmp)

	home := filepath.Join(tmp, ".agent-run")
	grokHome := filepath.Join(tmp, "grok-home")
	workspace := tmp
	if err := os.MkdirAll(grokHome, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir grok home: %v\n", err)
		return 1
	}

	agentRun, llmMock, err := ensureBinaries(repoRoot, tmp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build binaries: %v\n", err)
		return 1
	}

	hook := realLikeChromeHook(30)
	env := append(os.Environ(),
		"AGENT_RUN_HOME="+home,
		"AGENT_RUNNER_CONFIG_HOME="+grokHome,
		"GROK_HOME="+grokHome,
		"LLM_MOCK_RUN_GROK_COMMAND="+hook,
		"PATH="+filepath.Dir(agentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	if scenario == "delayed-session" {
		delayedUUID := "dddddddd-dddd-dddd-dddd-dddddddddddd"
		go scheduleDelayedGrokSession(grokHome, workspace, delayedUUID, prompt, 15*time.Second)
	} else if scenario != "chrome-cancel" {
		fmt.Fprintf(os.Stderr, "unknown scenario %q (use chrome-cancel or delayed-session)\n", scenario)
		return 2
	}

	fmt.Printf("=== scenario: %s ===\n", scenario)
	fmt.Printf("temp=%s\n", tmp)
	fmt.Printf("agent-run=%s\n", agentRun)
	fmt.Printf("llm-mock-run-grok=%s\n", llmMock)
	fmt.Printf("GROK_HOME=%s (empty at start)\n", grokHome)
	fmt.Printf("flags: --keep-tty (same as web KeepTerminalAlive)\n\n")

	args := []string{
		"run",
		"--agent-runner", "grok-tty",
		"--keep-tty",
		"--session", "debug_grok_cancel",
		"--agent-runner-binary", llmMock,
		"--agent-runner-config-home", grokHome,
		prompt,
	}
	cmd := exec.Command(agentRun, args...)
	cmd.Dir = workspace
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)

	eventsPath := filepath.Join(home, "sessions", "grok-tty", "debug_grok_cancel", "events.jsonl")
	events, readErr := readEventsJSONL(eventsPath)

	fmt.Printf("\n=== run finished in %s (runErr=%v readErr=%v) ===\n", elapsed.Round(time.Millisecond), runErr, readErr)
	printEvents(events)

	ok := analyzeOutcome(scenario, events, elapsed)
	if !ok {
		return 1
	}
	switch scenario {
	case "delayed-session":
		fmt.Println("\nOK: delayed session streamed after discovery (no early cancel)")
	default:
		fmt.Println("\nOK: empty-home bind waited past chrome false-complete window (no early context canceled)")
	}
	return 0
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "agent-run")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod with cmd/agent-run from %s", wd)
		}
		dir = parent
	}
}

func ensureBinaries(repoRoot, cacheDir string) (agentRun, llmMock string, err error) {
	cache := filepath.Join(cacheDir, "bin-cache")
	if err := os.MkdirAll(cache, 0755); err != nil {
		return "", "", err
	}
	agentRun = filepath.Join(cache, "agent-run")
	llmMock = filepath.Join(cache, "llm-mock-run-grok")
	for _, spec := range []struct {
		out  string
		args []string
	}{
		{agentRun, []string{"build", "-o", agentRun, "./cmd/agent-run"}},
		{llmMock, []string{"build", "-o", llmMock, "./agent/llm/llm-mock/llm-mock-run-grok"}},
	} {
		if st, statErr := os.Stat(spec.out); statErr == nil && st.Size() > 0 {
			continue
		}
		cmd := exec.Command("go", spec.args...)
		cmd.Dir = repoRoot
		if out, buildErr := cmd.CombinedOutput(); buildErr != nil {
			return "", "", fmt.Errorf("go %v: %w\n%s", spec.args, buildErr, string(out))
		}
	}
	return agentRun, llmMock, nil
}

func realLikeChromeHook(holdSec int) string {
	return fmt.Sprintf(`sh -c 'printf "⎇ master worktree ~/.wrk +1\n#1 %s\n⠦ Starting session… 0.9s\n╭----------------------------------------------------------------------------╮\n│ ❯                                                                          │\n╰---------------------------------------------- Grok Build · always-approve -╯\nShift+Tab:mode  │  Ctrl+;:queue  │  Ctrl+.:shortcuts\n"; sleep %d'`, prompt, holdSec)
}

func scheduleDelayedGrokSession(grokHome, workspace, sessionUUID, userPrompt string, delay time.Duration) {
	time.Sleep(delay)
	abs, _ := filepath.Abs(workspace)
	enc := url.PathEscape(abs)
	dir := filepath.Join(grokHome, "sessions", enc, sessionUUID)
	_ = os.MkdirAll(dir, 0755)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	summary := fmt.Sprintf(`{"info":{"cwd":%q,"sessionId":%q,"openedAt":%q},"created_at":%q}`,
		abs, sessionUUID, now, now)
	_ = os.WriteFile(filepath.Join(dir, "summary.json"), []byte(summary), 0644)
	updates := fmt.Sprintf(`{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":%q}}
{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"DELAYED_SESSION_MARKER"}}
`, userPrompt)
	_ = os.WriteFile(filepath.Join(dir, "updates.jsonl"), []byte(updates), 0644)
	fmt.Fprintf(os.Stderr, "\n[debug] delayed grok session materialized at %s after %s\n\n", dir, delay)
}

func readEventsJSONL(path string) ([]types.AgentEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []types.AgentEvent
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var ev types.AgentEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, sc.Err()
}

func printEvents(events []types.AgentEvent) {
	fmt.Println("events.jsonl:")
	for i, ev := range events {
		line, _ := json.Marshal(ev)
		fmt.Printf("  [%d] %s\n", i, string(line))
	}
}

const discoveryMinWindow = 3 * time.Second

func analyzeOutcome(scenario string, events []types.AgentEvent, elapsed time.Duration) bool {
	var hasThink, hasCancelError, hasResolveError, hasAssistant, hasDelayedMarker bool
	var thinkAt, errAt time.Time
	for _, ev := range events {
		switch ev.Type {
		case types.ActionThink:
			if strings.Contains(ev.Text, resolveThinkText) {
				hasThink = true
				if ev.Timestamp > 0 {
					thinkAt = time.UnixMilli(ev.Timestamp)
				}
			}
		case types.ActionError:
			if strings.HasPrefix(ev.Text, resolveErrorPrefix) {
				hasResolveError = true
				if ev.Timestamp > 0 {
					errAt = time.UnixMilli(ev.Timestamp)
				}
				if strings.Contains(ev.Text, "context canceled") {
					hasCancelError = true
				}
			}
		case types.ActionMessage:
			if ev.Role == "assistant" {
				hasAssistant = true
			}
			if strings.Contains(ev.Text, "DELAYED_SESSION_MARKER") {
				hasDelayedMarker = true
			}
		}
	}
	thinkToError := time.Duration(0)
	if hasThink && !thinkAt.IsZero() && !errAt.IsZero() {
		thinkToError = errAt.Sub(thinkAt)
	}

	fmt.Println("\n=== analysis ===")
	fmt.Printf("  resolve think card:      %v\n", hasThink)
	fmt.Printf("  resolve error:           %v\n", hasResolveError)
	fmt.Printf("  context canceled error:  %v\n", hasCancelError)
	fmt.Printf("  assistant fallback:      %v\n", hasAssistant)
	fmt.Printf("  delayed marker streamed: %v\n", hasDelayedMarker)
	fmt.Printf("  total elapsed:           %s\n", elapsed.Round(time.Millisecond))
	if thinkToError > 0 {
		fmt.Printf("  think→error gap:         %s\n", thinkToError.Round(time.Millisecond))
	}

	chrome := []byte(`⎇ master worktree
#1 one word of France capital
⠦ Starting session… 0.9s
╭----------------------------------------------------------------------------╮
│ ❯                                                                          │
╰---------------------------------------------- Grok Build · always-approve -╯`)
	extracted := agenttty.ExtractAssistantTextFromSnapshot("grok-tty", chrome, prompt)
	fmt.Printf("  chrome extract len:      %d (nonzero if scrollback falsely completes turn)\n", len(strings.TrimSpace(extracted)))

	if !hasThink {
		fmt.Println("FAIL: missing resolve think event")
		return false
	}

	switch scenario {
	case "delayed-session":
		if hasCancelError && thinkToError < 10*time.Second {
			fmt.Printf("FAIL: early context canceled within %s after think\n", thinkToError.Round(time.Millisecond))
			return false
		}
		if !hasDelayedMarker {
			fmt.Println("FAIL: expected DELAYED_SESSION_MARKER in assistant message")
			return false
		}
		if hasAssistant && !hasDelayedMarker {
			fmt.Println("FAIL: scrollback assistant leaked into events")
			return false
		}
		return true
	default:
		if !hasResolveError {
			fmt.Println("FAIL: expected error starting with 'Cannot resolve session id:'")
			return false
		}
		if hasCancelError && thinkToError < discoveryMinWindow {
			fmt.Printf("FAIL: early context canceled at %s (chrome false-complete bug)\n", thinkToError.Round(time.Millisecond))
			return false
		}
		if hasAssistant {
			fmt.Println("FAIL: scrollback assistant leaked into events")
			return false
		}
		if elapsed < discoveryMinWindow {
			fmt.Printf("NOTE: empty-home bind may take up to 90s; elapsed only %s\n", elapsed.Round(time.Millisecond))
		}
		return true
	}
}