package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: fake-opencode <command> [ARGS]

A fake opencode CLI that mimics the opencode surfaces used by agent-pro tests.

Commands:
  run               execute a prompt
  models            list deterministic fake models
  -h,--help         show help
`

const runHelp = `
Usage: fake-opencode run [OPTIONS] [PROMPT]

Options:
  --format <format>              output format; json is supported
  --dir <dir>                    working directory (ignored, for compatibility)
  --model <model>                model name
  --session <id>                 session id
  --mock-config <path>           JSON mock config with events, hooks, and exit behavior
  --dangerously-skip-permissions accepted for compatibility
  -h,--help                      show help
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "fake-opencode: %v\n", err)
		if err, ok := err.(*exitError); ok {
			os.Exit(err.Code)
		}
		os.Exit(1)
	}
}

type exitError struct {
	Code int
}

func (e *exitError) Error() string {
	return fmt.Sprintf("mock exited with code %d", e.Code)
}

func handle(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}
	switch args[0] {
	case "run":
		return handleRun(args[1:])
	case "models":
		return handleModels(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func handleModels(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print("Usage: fake-opencode models\n")
		return nil
	}
	if len(args) > 0 {
		return fmt.Errorf("models does not accept arguments")
	}
	fmt.Println("openai/gpt-5")
	fmt.Println("openai/gpt-5-mini")
	return nil
}

func handleRun(args []string) error {
	var formatFlag *string
	var dirFlag *string
	var modelFlag *string
	var sessionFlag *string
	var mockConfigFlag *string
	var skipPermissionsFlag *bool

	remaining, err := flags.String("--format", &formatFlag).
		String("--dir", &dirFlag).
		String("--model", &modelFlag).
		String("--session", &sessionFlag).
		String("--mock-config", &mockConfigFlag).
		Bool("--dangerously-skip-permissions", &skipPermissionsFlag).
		Help("-h,--help", runHelp).
		Parse(args)
	if err != nil {
		return err
	}
	_ = dirFlag
	_ = skipPermissionsFlag

	if sessionFlag != nil && strings.TrimSpace(*sessionFlag) != "" {
		sessionID := strings.TrimSpace(*sessionFlag)
		sessionsRoot := resolveSessionsRoot()
		sessionPath := filepath.Join(sessionsRoot, sessionID)
		info, statErr := os.Stat(sessionPath)
		if statErr != nil || !info.IsDir() {
			errorEvent := map[string]any{
				"type":      "error",
				"sessionID": sessionID,
				"error": map[string]any{
					"name": "Session not found",
					"data": map[string]any{
						"message": "Session not found",
					},
				},
			}
			line, _ := json.Marshal(errorEvent)
			fmt.Println(string(line))
			fmt.Fprintf(os.Stderr, "Session not found: %s\n", sessionID)
			return &exitError{Code: 1}
		}
	}

	prompt := strings.Join(remaining, " ")
	if prompt == "" {
		prompt = "help with the task"
	}

	mockConfigPath := os.Getenv("FAKE_OPENCODE_MOCK_CONFIG")
	if mockConfigFlag != nil && strings.TrimSpace(*mockConfigFlag) != "" {
		mockConfigPath = strings.TrimSpace(*mockConfigFlag)
	}

	cfg := &mockConfig{
		Runner:    "fake-opencode",
		SessionID: "fake-opencode-session",
		Model:     "openai/gpt-5",
		StdoutEvents: []map[string]any{
			{
				"type":      "text",
				"sessionID": "fake-opencode-session",
				"part": map[string]any{
					"id":   "p1",
					"type": "text",
					"text": "fake opencode answered",
				},
			},
		},
	}
	if mockConfigPath != "" {
		cfg, err = loadMockConfig(mockConfigPath)
		if err != nil {
			return err
		}
	}
	if cfg.Runner == "" {
		cfg.Runner = "fake-opencode"
	}
	if cfg.SessionID == "" && sessionFlag != nil {
		cfg.SessionID = strings.TrimSpace(*sessionFlag)
	}
	if cfg.SessionID == "" {
		cfg.SessionID = "fake-opencode-session"
	}
	if modelFlag != nil && strings.TrimSpace(*modelFlag) != "" {
		cfg.Model = strings.TrimSpace(*modelFlag)
	}
	if cfg.Model == "" {
		cfg.Model = "openai/gpt-5"
	}

	if err := cfg.fireHooks(hookTimingBeforeStart, prompt); err != nil {
		return err
	}
	if err := cfg.fireHooks(hookTimingBeforeStdout, prompt); err != nil {
		return err
	}

	jsonOutput := formatFlag == nil || strings.TrimSpace(*formatFlag) == "" || strings.TrimSpace(*formatFlag) == "json"
	if !jsonOutput {
		return fmt.Errorf("unsupported format: %s", strings.TrimSpace(*formatFlag))
	}
	for i, event := range cfg.StdoutEvents {
		line, err := json.Marshal(cfg.withSession(event))
		if err != nil {
			return fmt.Errorf("marshal stdout_events[%d]: %w", i, err)
		}
		fmt.Println(string(line))
		if cfg.DelayMS > 0 && i < len(cfg.StdoutEvents)-1 {
			time.Sleep(time.Duration(cfg.DelayMS) * time.Millisecond)
		}
	}

	if err := cfg.fireHooks(hookTimingAfterStdout, prompt); err != nil {
		return err
	}
	if cfg.Stderr != "" {
		fmt.Fprint(os.Stderr, cfg.Stderr)
		if !strings.HasSuffix(cfg.Stderr, "\n") {
			fmt.Fprintln(os.Stderr)
		}
	}
	if cfg.ExitCode != 0 {
		if err := cfg.fireHooks(hookTimingOnError, prompt); err != nil {
			return err
		}
		if err := cfg.fireHooks(hookTimingBeforeExit, prompt); err != nil {
			return err
		}
		return &exitError{Code: cfg.ExitCode}
	}
	if err := cfg.fireHooks(hookTimingBeforeExit, prompt); err != nil {
		return err
	}
	return nil
}

type hookTiming string

const (
	hookTimingBeforeStart  hookTiming = "before_start"
	hookTimingBeforeStdout hookTiming = "before_stdout"
	hookTimingAfterStdout  hookTiming = "after_stdout"
	hookTimingBeforeExit   hookTiming = "before_exit"
	hookTimingOnError      hookTiming = "on_error"
)

type mockConfig struct {
	Version          string           `json:"version"`
	Runner           string           `json:"runner"`
	SessionID        string           `json:"session_id"`
	Model            string           `json:"model"`
	DelayMS          int              `json:"delay_ms"`
	ExitCode         int              `json:"exit_code"`
	Stderr           string           `json:"stderr"`
	IgnoreHookErrors bool             `json:"ignore_hook_errors"`
	HookCommand      string           `json:"hook_command"`
	StdoutEvents     []map[string]any `json:"stdout_events"`
	Hooks            []mockHook       `json:"hooks"`
}

type mockHook struct {
	At      hookTiming      `json:"at"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

func loadMockConfig(path string) (*mockConfig, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("mock config path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read mock config %s: %w", path, err)
	}
	var cfg mockConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse mock config %s: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid mock config %s: %w", path, err)
	}
	return &cfg, nil
}

func (c *mockConfig) validate() error {
	if c.ExitCode < 0 {
		return fmt.Errorf("exit_code must be >= 0")
	}
	if c.DelayMS < 0 {
		return fmt.Errorf("delay_ms must be >= 0")
	}
	for i, event := range c.StdoutEvents {
		if eventType, _ := event["type"].(string); strings.TrimSpace(eventType) == "" {
			return fmt.Errorf("stdout_events[%d].type is required", i)
		}
	}
	for i, hook := range c.Hooks {
		if hook.At == "" {
			return fmt.Errorf("hooks[%d].at is required", i)
		}
		if !isValidHookTiming(hook.At) {
			return fmt.Errorf("hooks[%d].at is invalid: %s", i, hook.At)
		}
		if strings.TrimSpace(hook.Event) == "" {
			return fmt.Errorf("hooks[%d].event is required", i)
		}
	}
	if len(c.Hooks) > 0 && strings.TrimSpace(c.HookCommand) == "" {
		return fmt.Errorf("hook_command is required when hooks are configured")
	}
	return nil
}

func isValidHookTiming(at hookTiming) bool {
	switch at {
	case hookTimingBeforeStart, hookTimingBeforeStdout, hookTimingAfterStdout, hookTimingBeforeExit, hookTimingOnError:
		return true
	default:
		return false
	}
}

func (c *mockConfig) withSession(event map[string]any) map[string]any {
	out := make(map[string]any, len(event)+1)
	for k, v := range event {
		out[k] = v
	}
	if _, ok := out["sessionID"]; !ok {
		out["sessionID"] = c.SessionID
	}
	return out
}

func (c *mockConfig) fireHooks(at hookTiming, prompt string) error {
	for _, hook := range c.Hooks {
		if hook.At != at {
			continue
		}
		if err := c.fireHook(hook, prompt); err != nil {
			if c.IgnoreHookErrors {
				fmt.Fprintf(os.Stderr, "fake-opencode: hook %s failed: %v\n", hook.Event, err)
				continue
			}
			return err
		}
	}
	return nil
}

func (c *mockConfig) fireHook(hook mockHook, prompt string) error {
	payload, err := c.hookPayload(hook, prompt)
	if err != nil {
		return err
	}
	command := strings.ReplaceAll(c.HookCommand, "{{event}}", shellQuote(hook.Event))
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = bytes.NewReader(payload)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("hook %s: %w: %s", hook.Event, err, msg)
		}
		return fmt.Errorf("hook %s: %w", hook.Event, err)
	}
	return nil
}

func (c *mockConfig) hookPayload(hook mockHook, prompt string) ([]byte, error) {
	payload := map[string]any{}
	if len(hook.Payload) > 0 && string(hook.Payload) != "null" {
		if err := json.Unmarshal(hook.Payload, &payload); err != nil {
			return nil, fmt.Errorf("hook %s payload must be an object: %w", hook.Event, err)
		}
	}
	setDefault := func(key string, value any) {
		if _, ok := payload[key]; !ok && value != "" {
			payload[key] = value
		}
	}
	setDefault("event", hook.Event)
	setDefault("at", string(hook.At))
	setDefault("runner", c.Runner)
	setDefault("session_id", c.SessionID)
	setDefault("model", c.Model)
	setDefault("prompt", prompt)
	return json.Marshal(payload)
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func resolveSessionsRoot() string {
	configDir := os.Getenv("OPENCODE_CONFIG_DIR")
	if configDir != "" {
		return filepath.Join(configDir, "sessions")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode", "sessions")
}
