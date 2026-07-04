package run

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
)

// RunGrokOptions configures optional behavior for RunGrok.
type RunGrokOptions struct {
	MockEventsPreset string // --mock-events-preset; passed to mock server
	LogEventsPath    string // --log-events output path; passed to mock server as --agent-events-file
	LogHTTPPath      string // --log-http output path; passed to mock server as --log-http
}

// RunGrok starts the mock server in the background, configures an isolated GROK_HOME,
// runs grok in the foreground, and tears down the mock when grok exits.
func RunGrok(grokArgs []string, opts RunGrokOptions) error {
	if opts.LogEventsPath != "" && !strings.HasSuffix(opts.LogEventsPath, ".jsonl") {
		return fmt.Errorf("--log-events path must end with .jsonl")
	}
	if opts.LogHTTPPath != "" && !strings.HasSuffix(opts.LogHTTPPath, ".jsonl") {
		return fmt.Errorf("--log-http path must end with .jsonl")
	}

	loaded, err := mockconfig.LoadMerged("")
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "llm-mock-run-grok-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	// Reserve an ephemeral port so the orchestrator works when the default 8080
	// range is occupied (common in dev/CI with parallel tests).
	freePort, err := reservePort()
	if err != nil {
		return fmt.Errorf("reserve mock port: %w", err)
	}
	loaded.Config.Port = freePort

	mergedConfigPath := filepath.Join(tmpDir, "llm-mock-config.json")
	mergedData, err := loaded.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal merged config: %w", err)
	}
	if err := os.WriteFile(mergedConfigPath, mergedData, 0644); err != nil {
		return fmt.Errorf("write merged config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockCmd, port, err := startMockServer(ctx, mergedConfigPath, opts.MockEventsPreset, opts.LogEventsPath, opts.LogHTTPPath)
	if err != nil {
		return err
	}
	defer func() {
		teardownStart := time.Now()
		runGrokDebugf("mock teardown: cancel context")
		cancel()
		if mockCmd.Process != nil {
			runGrokDebugf("mock teardown: SIGTERM pid=%d", mockCmd.Process.Pid)
			_ = mockCmd.Process.Signal(syscall.SIGTERM)
			waitStart := time.Now()
			waitErr := mockCmd.Wait()
			runGrokDebugf("mock teardown: Wait returned after %s err=%v", since(waitStart), waitErr)
		}
		runGrokDebugf("mock teardown done after %s", since(teardownStart))
	}()

	grokHome, _, err := resolveGrokHome(tmpDir)
	if err != nil {
		return err
	}

	if err := writeGrokConfigToml(filepath.Join(grokHome, "config.toml"), port); err != nil {
		return err
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d/v1", port)
	grokEnv := append(os.Environ(),
		"GROK_HOME="+grokHome,
		"XAI_API_KEY=sk-mock",
		"GROK_MODELS_BASE_URL="+baseURL,
	)

	// Announce isolated home on stderr so test harnesses can locate session events.
	fmt.Fprintf(os.Stderr, "GROK_HOME=%s\n", grokHome)

	grokCmd, usedHook, err := buildGrokCommand(grokArgs)
	if err != nil {
		return err
	}
	grokCmd.Env = grokEnv
	workDir, err := workDirForSessionEncoding(".")
	if err != nil {
		return err
	}
	grokCmd.Dir = workDir
	grokCmd.Stdin = os.Stdin
	grokCmd.Stdout = os.Stdout
	grokCmd.Stderr = os.Stderr

	runGrokDebugf("starting grok (hook=%v) args=%q workDir=%q grokHome=%q mockPort=%d",
		usedHook, grokArgs, workDir, grokHome, port)
	grokStart := time.Now()
	runErr := grokCmd.Run()
	grokElapsed := since(grokStart)
	if runErr != nil {
		runGrokDebugf("grok exited with error after %s: %v", grokElapsed, runErr)
	} else {
		runGrokDebugf("grok exited ok after %s", grokElapsed)
	}
	if !usedHook {
		runGrokDebugf("post-grok session state: %s", describeSessionRoots(grokHome))
		mirrorStart := time.Now()
		// Real grok may flush session files asynchronously after exit.
		if err := waitAndMirrorSessions(grokHome, workDir); err != nil {
			fmt.Fprintf(os.Stderr, "llm-mock: mirror sessions: %v\n", err)
			runGrokDebugf("waitAndMirrorSessions failed after %s: %v", since(mirrorStart), err)
		} else if !mirroredSessionsReady(grokHome, workDir) {
			if sessionsHaveAnyEvents(grokHome) {
				fmt.Fprintf(os.Stderr, "llm-mock: mirror sessions: not ready for %s\n", workDir)
				runGrokDebugf("mirroredSessionsReady=false after %s workDir=%q enc=%q grokEnc=%q (events exist, mirror failed)",
					since(mirrorStart), workDir, url.PathEscape(workDir), grokSessionEncoding(workDir))
			} else {
				runGrokDebugf("mirroredSessionsReady=false after %s workDir=%q — no events to mirror (expected)",
					since(mirrorStart), workDir)
			}
		} else {
			runGrokDebugf("mirroredSessionsReady=true after %s", since(mirrorStart))
		}
		runGrokDebugf("post-mirror session state: %s", describeSessionRoots(grokHome))
	} else {
		runGrokDebugf("skipping waitAndMirrorSessions (LLM_MOCK_RUN_GROK_COMMAND hook)")
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("run grok: %w", runErr)
	}
	return nil
}

func resolveGrokHome(tmpDir string) (home string, cleanup bool, err error) {
	if v := os.Getenv("LLM_MOCK_GROK_HOME"); v != "" {
		if err := os.MkdirAll(v, 0755); err != nil {
			return "", false, fmt.Errorf("mkdir grok home: %w", err)
		}
		return v, false, nil
	}
	home = filepath.Join(tmpDir, "grok-home")
	if err := os.MkdirAll(home, 0755); err != nil {
		return "", false, fmt.Errorf("mkdir grok home: %w", err)
	}
	return home, true, nil
}

func buildGrokCommand(grokArgs []string) (*exec.Cmd, bool, error) {
	if hook := os.Getenv("LLM_MOCK_RUN_GROK_COMMAND"); hook != "" {
		return exec.Command("sh", "-c", hook), true, nil
	}
	if len(grokArgs) == 0 {
		return exec.Command("grok"), false, nil
	}
	return exec.Command("grok", grokArgs...), false, nil
}

func waitAndMirrorSessions(grokHome, workDir string) error {
	sessionsRoot := filepath.Join(grokHome, "sessions")
	if _, err := os.Stat(sessionsRoot); os.IsNotExist(err) {
		runGrokDebugf("waitAndMirrorSessions: no %s — return immediately", sessionsRoot)
		return nil
	}
	runGrokDebugf("waitAndMirrorSessions: enter workDir=%q enc=%q grokEnc=%q %s",
		workDir, url.PathEscape(workDir), grokSessionEncoding(workDir), describeSessionRoots(grokHome))
	loopStart := time.Now()
	deadline := time.Now().Add(60 * time.Second)
	iteration := 0
	for {
		iteration++
		if err := mirrorSessionsForWorkDir(grokHome, workDir); err != nil {
			runGrokDebugf("waitAndMirrorSessions: mirrorSessionsForWorkDir error on iter %d: %v", iteration, err)
			return err
		}
		if mirroredSessionsReady(grokHome, workDir) {
			runGrokDebugf("waitAndMirrorSessions: mirroredSessionsReady at %s iter=%d",
				since(loopStart), iteration)
			return nil
		}
		if !sessionsHaveAnyEvents(grokHome) {
			runGrokDebugf("waitAndMirrorSessions: no events.jsonl under any session — return immediately (iter=%d elapsed=%s)",
				iteration, since(loopStart))
			return nil
		}
		if iteration == 1 || iteration%25 == 0 {
			runGrokDebugf("waitAndMirrorSessions: iter %d events exist but not mirrored for workDir=%q (elapsed %s) %s",
				iteration, workDir, since(loopStart), describeSessionRoots(grokHome))
		}
		if time.Now().After(deadline) {
			runGrokDebugf("waitAndMirrorSessions: deadline reached after %s iter=%d %s",
				since(loopStart), iteration, describeSessionRoots(grokHome))
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func startMockServer(ctx context.Context, configPath, mockEventsPreset, eventsFile, logHTTPPath string) (*exec.Cmd, int, error) {
	exe, err := mockServerExecutable()
	if err != nil {
		return nil, 0, err
	}

	serverArgs := []string{"--config", configPath}
	if mockEventsPreset != "" {
		serverArgs = append(serverArgs, "--mock-events-preset", mockEventsPreset)
	}
	if eventsFile != "" {
		serverArgs = append(serverArgs, "--agent-events-file", eventsFile)
	}
	if logHTTPPath != "" {
		serverArgs = append(serverArgs, "--log-http", logHTTPPath)
	}
	cmd := exec.CommandContext(ctx, exe, serverArgs...)
	cmd.Env = envWithout(os.Environ(), "LLM_MOCK_EVENTS_FILE")
	cmd.Stderr = io.Discard

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, 0, fmt.Errorf("mock stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, 0, fmt.Errorf("start mock server: %w", err)
	}

	port, err := readPort(stdout)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, 0, fmt.Errorf("read mock port: %w", err)
	}
	return cmd, port, nil
}

func readPort(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("mock server produced no stdout")
	}
	line := strings.TrimSpace(scanner.Text())
	line = strings.TrimPrefix(line, ":")
	port, err := strconv.Atoi(line)
	if err != nil {
		return 0, fmt.Errorf("parse port from %q: %w", line, err)
	}
	return port, nil
}

func mockServerExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	dir := filepath.Dir(exe)
	if filepath.Base(exe) == "llm-mock-run-grok" {
		sibling := filepath.Join(dir, "llm-mock")
		if _, err := os.Stat(sibling); err == nil {
			return sibling, nil
		}
	}
	return exe, nil
}

func reservePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected listener addr type %T", ln.Addr())
	}
	return tcpAddr.Port, nil
}

func envWithout(base []string, key string) []string {
	out := make([]string, 0, len(base))
	prefix := key + "="
	for _, kv := range base {
		if strings.HasPrefix(kv, prefix) {
			continue
		}
		out = append(out, kv)
	}
	return out
}

func writeGrokConfigToml(path string, port int) error {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d/v1", port)
	content := fmt.Sprintf(`[endpoints]
models_base_url = %q

[models]
default = "mock-model"

[model.mock-model]
model = "mock-model"
base_url = %q
api_key = "sk-mock"
name = "Mock Model"
api_backend = "chat_completions"
`, baseURL, baseURL)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write grok config.toml: %w", err)
	}
	return nil
}