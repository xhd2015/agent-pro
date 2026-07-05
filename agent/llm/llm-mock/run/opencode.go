package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockconfig"
)

// RunOpencodeOptions configures optional behavior for RunOpencode.
type RunOpencodeOptions struct {
	MockEventsPreset string // --mock-events-preset; passed to mock server
	LogEventsPath    string // --log-events output path; passed to mock server as --agent-events-file
	LogHTTPPath      string // --log-http output path; passed to mock server as --log-http
}

// RunOpencode starts the mock server in the background, configures an isolated opencode env,
// runs opencode in the foreground, and tears down the mock when opencode exits.
func RunOpencode(opencodeArgs []string, opts RunOpencodeOptions) error {
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

	tmpDir, err := os.MkdirTemp("", "llm-mock-run-opencode-*")
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

	mockCmd, port, err := startMockServer(ctx, mergedConfigPath, opts.MockEventsPreset, opts.LogEventsPath, opts.LogHTTPPath)
	if err != nil {
		return err
	}
	defer func() {
		teardownStart := time.Now()
		runOpencodeDebugf("mock teardown: cancel context")
		cancel()
		if mockCmd.Process != nil {
			runOpencodeDebugf("mock teardown: SIGTERM pid=%d", mockCmd.Process.Pid)
			_ = mockCmd.Process.Signal(syscall.SIGTERM)
			waitStart := time.Now()
			waitErr := mockCmd.Wait()
			runOpencodeDebugf("mock teardown: Wait returned after %s err=%v", since(waitStart), waitErr)
		}
		runOpencodeDebugf("mock teardown done after %s", since(teardownStart))
	}()

	opencodeConfigDir, err := resolveOpencodeConfigDir(tmpDir)
	if err != nil {
		return err
	}

	opencodeHome, err := resolveOpencodeHome(tmpDir)
	if err != nil {
		return err
	}

	configContent, err := buildOpencodeConfigContent(port)
	if err != nil {
		return err
	}

	opencodeEnv := append(os.Environ(),
		"HOME="+opencodeHome,
		"OPENCODE_CONFIG_DIR="+opencodeConfigDir,
		"OPENCODE_CONFIG_CONTENT="+configContent,
		"OPENCODE_DISABLE_PROJECT_CONFIG=1",
		"OPENCODE_PURE=1",
		"OPENCODE_DISABLE_MODELS_FETCH=1",
		"OPENCODE_AUTH_CONTENT={}",
	)

	fmt.Fprintf(os.Stderr, "OPENCODE_CONFIG_DIR=%s\n", opencodeConfigDir)
	fmt.Fprintf(os.Stderr, "HOME=%s\n", opencodeHome)
	fmt.Fprintf(os.Stderr, "OPENCODE_CONFIG_CONTENT=%s\n", configContent)

	opencodeCmd, usedHook, err := buildOpencodeCommand(opencodeArgs)
	if err != nil {
		return err
	}
	opencodeCmd.Env = opencodeEnv
	opencodeCmd.Stdin = os.Stdin
	opencodeCmd.Stdout = os.Stdout
	opencodeCmd.Stderr = os.Stderr

	runOpencodeDebugf("starting opencode (hook=%v) args=%q configDir=%q home=%q mockPort=%d",
		usedHook, opencodeArgs, opencodeConfigDir, opencodeHome, port)
	opencodeStart := time.Now()
	runErr := opencodeCmd.Run()
	opencodeElapsed := since(opencodeStart)
	if runErr != nil {
		runOpencodeDebugf("opencode exited with error after %s: %v", opencodeElapsed, runErr)
	} else {
		runOpencodeDebugf("opencode exited ok after %s", opencodeElapsed)
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("run opencode: %w", runErr)
	}
	return nil
}

func resolveOpencodeConfigDir(tmpDir string) (string, error) {
	if v := os.Getenv("LLM_MOCK_OPENCODE_CONFIG_DIR"); v != "" {
		if err := os.MkdirAll(v, 0755); err != nil {
			return "", fmt.Errorf("mkdir opencode config dir: %w", err)
		}
		return v, nil
	}
	dir := filepath.Join(tmpDir, "opencode-config")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdir opencode config dir: %w", err)
	}
	return dir, nil
}

func resolveOpencodeHome(tmpDir string) (string, error) {
	if v := os.Getenv("LLM_MOCK_OPENCODE_HOME"); v != "" {
		if err := os.MkdirAll(v, 0755); err != nil {
			return "", fmt.Errorf("mkdir opencode home: %w", err)
		}
		return v, nil
	}
	home := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(home, 0755); err != nil {
		return "", fmt.Errorf("mkdir opencode home: %w", err)
	}
	return home, nil
}

func buildOpencodeCommand(opencodeArgs []string) (*exec.Cmd, bool, error) {
	if hook := os.Getenv("LLM_MOCK_RUN_OPENCODE_COMMAND"); hook != "" {
		return exec.Command("sh", "-c", hook), true, nil
	}
	if len(opencodeArgs) == 0 {
		return exec.Command("opencode"), false, nil
	}
	return exec.Command("opencode", opencodeArgs...), false, nil
}

func buildOpencodeConfigContent(port int) (string, error) {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d/v1", port)
	cfg := map[string]any{
		"formatter": false,
		"lsp":       false,
		"model":     "llm-mock/mock-model",
		"provider": map[string]any{
			"llm-mock": map[string]any{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "LLM Mock",
				"options": map[string]any{
					"apiKey":  "sk-mock",
					"baseURL": baseURL,
				},
				"models": map[string]any{
					"mock-model": map[string]any{
						"name": "Mock Model",
						"limit": map[string]any{
							"context": 100000,
							"output":  10000,
						},
					},
				},
			},
		},
		"permission": map[string]any{
			"question":   "allow",
			"plan_enter": "allow",
			"plan_exit":  "allow",
		},
		"agent": map[string]any{
			"title": map[string]any{
				"disable": true,
			},
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal opencode config content: %w", err)
	}
	return string(data), nil
}