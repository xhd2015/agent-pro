package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/server"
)

// RunCommandCodeOptions configures optional behavior for RunCommandCode.
type RunCommandCodeOptions struct {
	MockEventsPreset string // --mock-events-preset; passed to mock server
	MockEventsFile   string // --mock-events-file; AgentEvent JSONL for genQueue
	LogEventsPath    string // --log-events output path; passed to mock server as --agent-events-file
	LogHTTPPath      string // --log-http output path; passed to mock server as --log-http
}

// RunCommandCode starts the mock server in-process (goroutine), sets sandbox env vars,
// isolates HOME to a temp dir, runs cmd in the foreground, and tears down on exit.
func RunCommandCode(cmdArgs []string, opts RunCommandCodeOptions) error {
	if hook := strings.TrimSpace(os.Getenv("LLM_MOCK_RUN_COMMANDCODE_COMMAND")); hook != "" {
		return runCommandCodeHookCommand(hook)
	}

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

	tmpDir, err := os.MkdirTemp("", "llm-mock-run-commandcode-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

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

	srv, err := server.Start(ctx, server.Options{
		Config:           loaded,
		LogHTTPPath:      opts.LogHTTPPath,
		AgentEventsFile:  opts.LogEventsPath,
		MockEventsPreset: opts.MockEventsPreset,
		MockEventsFile:   opts.MockEventsFile,
	})
	if err != nil {
		return err
	}
	defer srv.Close()

	port := srv.Port()

	// Isolate HOME so ~/.commandcode from real usage is not touched.
	isolatedHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(isolatedHome, 0700); err != nil {
		return fmt.Errorf("mkdir isolated home: %w", err)
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	ccEnv := append(os.Environ(),
		"HOME="+isolatedHome,
		"COMMANDCODE_SANDBOX=true",
		"COMMANDCODE_API_URL="+baseURL,
		"COMMAND_CODE_API_KEY=sk-mock",
	)

	ccCmd, usedHook, err := buildCommandCodeCommand(cmdArgs)
	if err != nil {
		return err
	}
	ccCmd.Env = ccEnv
	ccCmd.Stdin = os.Stdin
	ccCmd.Stdout = os.Stdout
	ccCmd.Stderr = os.Stderr

	_ = usedHook
	runCommandCodeDebugf("starting commandcode (hook=%v) args=%q mockPort=%d",
		usedHook, cmdArgs, port)
	if err := ccCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("run commandcode: %w", err)
	}
	return nil
}

func runCommandCodeHookCommand(hook string) error {
	ccCmd := exec.Command("sh", "-c", hook)
	ccCmd.Stdin = os.Stdin
	ccCmd.Stdout = os.Stdout
	ccCmd.Stderr = os.Stderr

	runCommandCodeDebugf("starting commandcode hook")
	runErr := ccCmd.Run()
	if runErr != nil {
		runCommandCodeDebugf("commandcode hook exited with error: %v", runErr)
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("run commandcode hook: %w", runErr)
	}
	runCommandCodeDebugf("commandcode hook exited ok")
	return nil
}

func buildCommandCodeCommand(cmdArgs []string) (*exec.Cmd, bool, error) {
	if hook := os.Getenv("LLM_MOCK_RUN_COMMANDCODE_COMMAND"); hook != "" {
		return exec.Command("sh", "-c", hook), true, nil
	}
	args := append([]string{}, cmdArgs...)
	args = append(args, "--yolo")
	return exec.Command("cmd", args...), false, nil
}

func runCommandCodeDebugEnabled() bool {
	v := os.Getenv("LLM_MOCK_RUN_COMMANDCODE_DEBUG")
	return v == "1" || v == "true" || v == "yes"
}

func runCommandCodeDebugf(format string, args ...any) {
	if !runCommandCodeDebugEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "llm-mock[run-commandcode]: "+format+"\n", args...)
}
