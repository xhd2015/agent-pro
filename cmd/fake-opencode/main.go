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

	types "github.com/xhd2015/agent-pro/agent/event/types"
	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
	"github.com/xhd2015/agent-pro/pkgs/fake-agent/events"

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
  --seed <int>                   random seed for deterministic output
  --mock-config <path>           JSON mock config with events, hooks, and exit behavior
  --plugin <path>                plugin file to load (can be specified multiple times)
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
	var seedFlag *int64
	var pluginFlags []string

	remaining, err := flags.String("--format", &formatFlag).
		String("--dir", &dirFlag).
		String("--model", &modelFlag).
		String("--session", &sessionFlag).
		String("--mock-config", &mockConfigFlag).
		Bool("--dangerously-skip-permissions", &skipPermissionsFlag).
		Int("--seed", &seedFlag).
		StringSlice("--plugin", &pluginFlags).
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
			errorEvent := opencode_types.Event{
				Type:      opencode_types.EvtError,
				SessionID: sessionID,
				Error: &opencode_types.ErrorDetail{
					Name: "Session not found",
					Data: &opencode_types.ErrorData{
						Message: "Session not found",
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

	var cfg *mockConfig
	if mockConfigPath != "" {
		var err error
		cfg, err = loadMockConfig(mockConfigPath)
		if err != nil {
			return err
		}
	} else {
		var seed int64
		if seedFlag != nil {
			seed = *seedFlag
		} else {
			seed = time.Now().UnixNano()
		}
		agentEvents := events.GenerateEvents(seed, prompt)
		cfg = &mockConfig{
			Runner:       "fake-opencode",
			SessionID:    fmt.Sprintf("session_%d", seed),
			Model:        "openai/gpt-5",
			StdoutEvents: opencode_types.ToOpencode(agentEvents, ""),
			AgentEvents:  agentEvents,
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

	allPlugins := mergePlugins(cfg.Plugins, pluginFlags)

	configHome := resolveConfigHome()
	pluginsDir := filepath.Join(configHome, "plugins")
	if entries, err := os.ReadDir(pluginsDir); err == nil {
		seen := make(map[string]bool)
		for _, p := range allPlugins {
			seen[p] = true
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) != ".ts" {
				continue
			}
			absPath, err := filepath.Abs(filepath.Join(pluginsDir, entry.Name()))
			if err != nil {
				continue
			}
			if seen[absPath] {
				continue
			}
			seen[absPath] = true
			allPlugins = append(allPlugins, absPath)
		}
	}
	var pluginRunner string
	if len(allPlugins) > 0 {
		var err error
		pluginRunner, err = generatePluginRunner(allPlugins)
		if err != nil {
			return fmt.Errorf("generate plugin runner: %w", err)
		}
		defer os.Remove(pluginRunner)
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

	if pluginRunner != "" {
		firePluginEvent(pluginRunner, "session.created", map[string]any{
			"type":       "session.created",
			"session_id": cfg.SessionID,
			"runner":     cfg.Runner,
			"model":      cfg.Model,
		})
	}

	stdoutIdx := 0
	for i, ae := range cfg.AgentEvents {
		if ae.Type == types.ActionSleep {
			time.Sleep(time.Duration(ae.DelayMs) * time.Millisecond)
			continue
		}

		if stdoutIdx >= len(cfg.StdoutEvents) {
			break
		}
		event := cfg.StdoutEvents[stdoutIdx]
		stdoutIdx++

		if ae.DelayMs > 0 {
			time.Sleep(time.Duration(ae.DelayMs) * time.Millisecond)
		}

		if pluginRunner != "" {
			evt := cfg.withSession(event)
			firePluginEvent(pluginRunner, string(evt.Type), map[string]any{
				"type":       string(evt.Type),
				"session_id": cfg.SessionID,
				"runner":     cfg.Runner,
				"model":      cfg.Model,
			})
		}
		line, err := json.Marshal(cfg.withSession(event))
		if err != nil {
			return fmt.Errorf("marshal stdout_events[%d]: %w", i, err)
		}
		fmt.Println(string(line))
		if cfg.DelayMS > 0 && stdoutIdx < len(cfg.StdoutEvents) {
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
		if pluginRunner != "" {
			firePluginEvent(pluginRunner, "session.error", map[string]any{
				"type":       "session.error",
				"session_id": cfg.SessionID,
				"runner":     cfg.Runner,
				"model":      cfg.Model,
			})
		}
		if err := cfg.fireHooks(hookTimingOnError, prompt); err != nil {
			return err
		}
		if err := cfg.fireHooks(hookTimingBeforeExit, prompt); err != nil {
			return err
		}
		return &exitError{Code: cfg.ExitCode}
	}
	if pluginRunner != "" {
		firePluginEvent(pluginRunner, "session.idle", map[string]any{
			"type":       "session.idle",
			"session_id": cfg.SessionID,
			"runner":     cfg.Runner,
			"model":      cfg.Model,
		})
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
	Version          string                 `json:"version"`
	Runner           string                 `json:"runner"`
	SessionID        string                 `json:"session_id"`
	Model            string                 `json:"model"`
	DelayMS          int                    `json:"delay_ms"`
	ExitCode         int                    `json:"exit_code"`
	Stderr           string                 `json:"stderr"`
	IgnoreHookErrors bool                   `json:"ignore_hook_errors"`
	HookCommand      string                 `json:"hook_command"`
	StdoutEventsRaw  json.RawMessage        `json:"stdout_events"`
	LLMEventsRaw     json.RawMessage        `json:"llm_events"`
	StdoutEvents     []opencode_types.Event `json:"-"`
	LLMEvents        []opencode_types.Event `json:"-"`
	AgentEvents      []types.AgentEvent     `json:"-"`
	Hooks            []mockHook             `json:"hooks"`
	Plugins          []string               `json:"plugins"`
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

	hasStdoutEvents := cfg.StdoutEventsRaw != nil && string(cfg.StdoutEventsRaw) != "null"
	hasLLMEvents := cfg.LLMEventsRaw != nil

	var rawEvents json.RawMessage
	if hasLLMEvents {
		rawEvents = cfg.LLMEventsRaw
	} else {
		rawEvents = cfg.StdoutEventsRaw
	}
	agentEvents, events, err := resolveEvents(rawEvents)
	if err != nil {
		return nil, err
	}
	cfg.AgentEvents = agentEvents
	cfg.StdoutEvents = events

	if hasStdoutEvents {
		fmt.Fprintf(os.Stderr, "Warning: stdout_events is deprecated, use llm_events instead\n")
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
		if strings.TrimSpace(string(event.Type)) == "" && !event.Done {
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
		c.HookCommand = "agent-hub hook notify --runner " + c.Runner + " --event {{event}}"
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

func (c *mockConfig) withSession(event opencode_types.Event) opencode_types.Event {
	if event.SessionID == "" {
		event.SessionID = c.SessionID
	}
	return event
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

func resolveConfigHome() string {
	if configDir := os.Getenv("OPENCODE_CONFIG_DIR"); configDir != "" {
		return configDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode")
}

func resolveEvents(raw json.RawMessage) ([]types.AgentEvent, []opencode_types.Event, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil, nil
	}

	var rawEvents []json.RawMessage
	if err := json.Unmarshal(raw, &rawEvents); err != nil {
		return nil, nil, fmt.Errorf("parse events array: %w", err)
	}

	var agentEvents []types.AgentEvent
	var events []opencode_types.Event
	for i, rawEvt := range rawEvents {
		var ae types.AgentEvent
		if err := json.Unmarshal(rawEvt, &ae); err != nil {
			return nil, nil, fmt.Errorf("parse event[%d]: %w", i, err)
		}
		if !isValidActionType(ae.Type) {
			return nil, nil, fmt.Errorf("events[%d]: unrecognized event type %q", i, ae.Type)
		}
		agentEvents = append(agentEvents, ae)
		events = append(events, opencode_types.ToOpencode([]types.AgentEvent{ae}, "")...)
	}

	return agentEvents, events, nil
}


func isValidActionType(t types.ActionType) bool {
	switch t {
	case types.ActionThink, types.ActionToolCall, types.ActionMessage,
		types.ActionError, types.ActionDone, types.ActionStepStart, types.ActionStepFinish,
		types.ActionSleep:
		return true
	}
	return false
}

func mergePlugins(configPlugins, cliPlugins []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, p := range configPlugins {
		if p == "" {
			continue
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		result = append(result, p)
	}
	for _, p := range cliPlugins {
		if p == "" {
			continue
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		result = append(result, p)
	}
	return result
}

func generatePluginRunner(plugins []string) (string, error) {
	f, err := os.CreateTemp("", "fake-opencode-plugin-*.ts")
	if err != nil {
		return "", err
	}

	for i, plugin := range plugins {
		absPath, err := filepath.Abs(plugin)
		if err != nil {
			f.Close()
			os.Remove(f.Name())
			return "", fmt.Errorf("resolve plugin path %s: %w", plugin, err)
		}
		fmt.Fprintf(f, "import { AgentHubPlugin } from '%s';\n", absPath)
		fmt.Fprintf(f, "const plugin%d = await AgentHubPlugin({ project: {}, client: {}, $: process.env, directory: process.cwd() });\n", i)
	}
	fmt.Fprint(f, `let input = '';
for await (const chunk of process.stdin) { input += typeof chunk === 'string' ? chunk : new TextDecoder().decode(chunk); }
const event = JSON.parse(input);
`)
	for i := range plugins {
		fmt.Fprintf(f, "const handler%d = plugin%d[event.type];\n", i, i)
		fmt.Fprintf(f, "if (handler%d) {\n", i)
		fmt.Fprintf(f, "  try { await handler%d(event); } catch(e) { console.error(e); }\n", i)
		fmt.Fprintf(f, "}\n")
	}

	f.Close()
	return f.Name(), nil
}

func firePluginEvent(pluginRunner string, eventType string, payload any) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fake-opencode: plugin event marshal failed: %v\n", err)
		return
	}

	cmd := exec.Command("bun", pluginRunner)
	cmd.Stdin = bytes.NewReader(payloadJSON)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_ = cmd.Run()
	msg := strings.TrimSpace(stderr.String())
	if msg != "" {
		fmt.Fprint(os.Stderr, msg)
		if !strings.HasSuffix(msg, "\n") {
			fmt.Fprintln(os.Stderr)
		}
	}
}

