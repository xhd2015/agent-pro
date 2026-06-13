package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	codex_types "github.com/xhd2015/agent-pro/agent/event/codex_types"
	fakeagent "github.com/xhd2015/agent-pro/pkgs/fake-agent"
	faketoolexec "github.com/xhd2015/agent-pro/pkgs/fake-agent/fake-tool-exec"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: fake-codex <command> [ARGS]

A fake codex CLI that mimics codex behavior without real LLM calls.

Commands:
  exec              execute a prompt (mimics codex exec)
  -h,--help         show help

Run fake-codex exec -h for exec-specific options.
`

const execHelp = `
Usage: fake-codex exec [OPTIONS] [PROMPT]

Mimics codex exec --json output with randomly generated events.

Options:
  --json                          output JSON events (default: human-friendly)
  --cd <dir>                      working directory (ignored, for compatibility)
  --sandbox <mode>                sandbox mode (ignored, for compatibility)
  --model <model>                 model name (ignored, for compatibility)
  --script <path>                 JSON script with exact events to emit
  --mock-config <path>            JSON mock config with events, hooks, and exit behavior
  --skip-git-repo-check           skip git check (ignored, for compatibility)
  --seed <int>                    random seed for deterministic output
  --delay <ms>                    delay in ms between output lines (default: 0)
  -h,--help                       show help
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "fake-codex: %v\n", err)
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
	return fmt.Sprintf("script exited with code %d", e.Code)
}

func handle(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}

	switch args[0] {
	case "exec":
		return handleExec(args[1:])
	default:
		return fmt.Errorf("unknown command: %s (only 'exec' is currently supported)", args[0])
	}
}

func handleExec(args []string) error {
	var dirFlag *string
	var sandboxFlag *string
	var modelFlag *string
	var scriptFlag *string
	var mockConfigFlag *string
	var jsonFlag *bool
	var skipGitFlag *bool
	var seedFlag *int64
	var delayFlag *int

	remaining, err := flags.String("--cd", &dirFlag).
		String("--sandbox", &sandboxFlag).
		String("--model", &modelFlag).
		String("--script", &scriptFlag).
		String("--mock-config", &mockConfigFlag).
		Bool("--json", &jsonFlag).
		Bool("--skip-git-repo-check", &skipGitFlag).
		Int("--seed", &seedFlag).
		Int("--delay", &delayFlag).
		Help("-h,--help", execHelp).
		Parse(args)
	if err != nil {
		return err
	}

	prompt := strings.Join(remaining, " ")
	if prompt == "" {
		prompt = "help with the task"
	}

	var seed int64
	if seedFlag != nil {
		seed = *seedFlag
	} else {
		seed = time.Now().UnixNano()
	}

	var delay time.Duration
	if delayFlag != nil && *delayFlag > 0 {
		delay = time.Duration(*delayFlag) * time.Millisecond
	}

	var events []fakeagent.Event
	var exitCode int
	var scriptedStderr string
	var mockConfig *mockConfig
	scriptPath := os.Getenv("FAKE_CODEX_SCRIPT")
	if scriptFlag != nil && strings.TrimSpace(*scriptFlag) != "" {
		scriptPath = strings.TrimSpace(*scriptFlag)
	}
	mockConfigPath := os.Getenv("FAKE_CODEX_MOCK_CONFIG")
	if mockConfigFlag != nil && strings.TrimSpace(*mockConfigFlag) != "" {
		mockConfigPath = strings.TrimSpace(*mockConfigFlag)
	}
	if mockConfigPath != "" {
		cfg, err := loadMockConfig(mockConfigPath)
		if err != nil {
			return err
		}
		mockConfig = cfg
		events = cfg.StdoutEvents
		exitCode = cfg.ExitCode
		scriptedStderr = cfg.Stderr
		if delayFlag == nil && cfg.DelayMS > 0 {
			delay = time.Duration(cfg.DelayMS) * time.Millisecond
		}
	} else if scriptPath != "" {
		script, err := fakeagent.LoadScript(scriptPath)
		if err != nil {
			return err
		}
		events = script.Events
		exitCode = script.ExitCode
		scriptedStderr = script.Stderr
		if delayFlag == nil && script.DelayMS > 0 {
			delay = time.Duration(script.DelayMS) * time.Millisecond
		}
	} else {
		gen := fakeagent.NewGenerator(seed)
		events = gen.GenerateSession(prompt)
	}

	if os.Getenv("CODEX_THREAD_ID") == "" {
		threadID := ""
		if mockConfig != nil && mockConfig.SessionID != "" {
			threadID = mockConfig.SessionID
		}
		if threadID == "" {
			threadID = fmt.Sprintf("codex_%d", time.Now().UnixNano())
		}
		os.Setenv("CODEX_THREAD_ID", threadID)
	}

	if mockConfig != nil {
		for i := range events {
			processFakeCodexEvent(&events[i])
		}
	}

	jsonOutput := jsonFlag != nil && *jsonFlag

	if mockConfig != nil {
		if err := mockConfig.fireHooks(hookTimingBeforeStart, prompt); err != nil {
			return err
		}
		if err := mockConfig.fireHooks(hookTimingBeforeStdout, prompt); err != nil {
			return err
		}
	}

	if jsonOutput {
		lines, err := fakeagent.FormatCodexEvents(events)
		if err != nil {
			return fmt.Errorf("format events: %w", err)
		}
		for i, line := range lines {
			fmt.Println(line)
			if delay > 0 && i < len(lines)-1 {
				time.Sleep(delay)
			}
		}
	} else {
		text := fakeagent.FormatCodexEventsText(events)
		if delay > 0 {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				fmt.Println(line)
				time.Sleep(delay)
			}
		} else {
			fmt.Print(text)
		}
	}
	if mockConfig != nil {
		if err := mockConfig.fireHooks(hookTimingAfterStdout, prompt); err != nil {
			return err
		}
	}
	if scriptedStderr != "" {
		fmt.Fprint(os.Stderr, scriptedStderr)
		if !strings.HasSuffix(scriptedStderr, "\n") {
			fmt.Fprintln(os.Stderr)
		}
	}
	if exitCode != 0 {
		if mockConfig != nil {
			if err := mockConfig.fireHooks(hookTimingOnError, prompt); err != nil {
				return err
			}
			if err := mockConfig.fireHooks(hookTimingBeforeExit, prompt); err != nil {
				return err
			}
		}
		return &exitError{Code: exitCode}
	}
	if mockConfig != nil {
		if err := mockConfig.fireHooks(hookTimingBeforeExit, prompt); err != nil {
			return err
		}
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
	Version            string            `json:"version"`
	Runner             string            `json:"runner"`
	SessionID          string            `json:"session_id"`
	Model              string            `json:"model"`
	DelayMS            int               `json:"delay_ms"`
	ExitCode           int               `json:"exit_code"`
	Stderr             string            `json:"stderr"`
	IgnoreHookErrors   bool              `json:"ignore_hook_errors"`
	HookCommand        string            `json:"hook_command"`
	StdoutEventsRaw    json.RawMessage   `json:"stdout_events"`
	LLMEventsRaw       json.RawMessage   `json:"llm_events"`
	StdoutEvents       []fakeagent.Event `json:"-"`
	LLMEvents          []fakeagent.Event `json:"-"`
	Hooks              []mockHook        `json:"hooks"`
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

	var events []fakeagent.Event
	if hasLLMEvents {
		events, err = resolveCodexLLMEvents(cfg.LLMEventsRaw)
		if err != nil {
			return nil, fmt.Errorf("resolve llm_events: %w", err)
		}
	} else {
		events, err = resolveCodexStdoutEvents(cfg.StdoutEventsRaw)
		if err != nil {
			return nil, fmt.Errorf("resolve stdout_events: %w", err)
		}
	}
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
		if event.Type == "" {
			return fmt.Errorf("stdout_events[%d].type is required", i)
		}
		if event.Item != nil && event.Item.Type == "" {
			return fmt.Errorf("stdout_events[%d].item.type is required", i)
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

func (c *mockConfig) fireHooks(at hookTiming, prompt string) error {
	for _, hook := range c.Hooks {
		if hook.At != at {
			continue
		}
		if err := c.fireHook(hook, prompt); err != nil {
			if c.IgnoreHookErrors {
				fmt.Fprintf(os.Stderr, "fake-codex: hook %s failed: %v\n", hook.Event, err)
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

func processFakeCodexEvent(event *fakeagent.Event) {
	if event.Item == nil {
		return
	}

	if event.Mock != nil {
		processMockEvent(event)
		return
	}

	switch event.Item.Type {
	case fakeagent.ItemCommandExecution:
		if event.Item.Command == "" {
			return
		}
		stdout, _, exitCode, _ := faketoolexec.ExecuteBash(event.Item.Command, "", nil)
		event.Item.AggregatedOutput = stdout
		event.Item.ExitCode = &exitCode
		event.Item.Status = "completed"
	case fakeagent.ItemFileChange:
		if event.Item.Status == "completed" && len(event.Item.Changes) > 0 {
			return
		}
		for _, change := range event.Item.Changes {
			faketoolexec.ExecuteWrite(change.Path, "content written by agent")
		}
		event.Item.Status = "completed"
	}
}

func processMockEvent(event *fakeagent.Event) {
	mock := event.Mock
	event.Mock = nil

	switch event.Item.Type {
	case fakeagent.ItemCommandExecution:
		event.Item.AggregatedOutput = mock.Output
		ec := 0
		if mock.ExitCode != nil {
			ec = *mock.ExitCode
		}
		event.Item.ExitCode = &ec
		event.Item.Status = "completed"
	case fakeagent.ItemFileChange:
		if len(mock.Changes) > 0 {
			var changes []fakeagent.FileChange
			for _, c := range mock.Changes {
				changes = append(changes, fakeagent.FileChange{Path: c.Path, Kind: c.Kind})
			}
			event.Item.Changes = changes
		}
		event.Item.Status = "completed"
	}
}

func resolveCodexStdoutEvents(raw json.RawMessage) ([]fakeagent.Event, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var rawEvents []json.RawMessage
	if err := json.Unmarshal(raw, &rawEvents); err != nil {
		return nil, fmt.Errorf("parse stdout_events array: %w", err)
	}

	var result []fakeagent.Event
	for i, rawEvt := range rawEvents {
		var probe map[string]any
		if err := json.Unmarshal(rawEvt, &probe); err != nil {
			return nil, fmt.Errorf("parse stdout_events[%d]: %w", i, err)
		}

		if _, hasItem := probe["item"]; hasItem {
			var evt fakeagent.Event
			if err := json.Unmarshal(rawEvt, &evt); err != nil {
				return nil, fmt.Errorf("parse stdout_events[%d] as codex event: %w", i, err)
			}
			result = append(result, evt)
		} else if _, hasMessage := probe["message"]; hasMessage {
			var evt fakeagent.Event
			if err := json.Unmarshal(rawEvt, &evt); err != nil {
				return nil, fmt.Errorf("parse stdout_events[%d] as codex event: %w", i, err)
			}
			result = append(result, evt)
		} else {
			var ae types.AgentEvent
			if err := json.Unmarshal(rawEvt, &ae); err != nil {
				return nil, fmt.Errorf("parse stdout_events[%d] as agent event: %w", i, err)
			}
			result = append(result, codex_types.ToCodex([]types.AgentEvent{ae})...)
		}
	}

	return result, nil
}

func resolveCodexLLMEvents(raw json.RawMessage) ([]fakeagent.Event, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var rawEvents []json.RawMessage
	if err := json.Unmarshal(raw, &rawEvents); err != nil {
		return nil, fmt.Errorf("parse llm_events array: %w", err)
	}

	var result []fakeagent.Event
	for i, rawEvt := range rawEvents {
		var ae types.AgentEvent
		if err := json.Unmarshal(rawEvt, &ae); err != nil {
			return nil, fmt.Errorf("parse llm_events[%d] as agent event: %w", i, err)
		}
		if !isValidActionType(ae.Type) {
			return nil, fmt.Errorf("llm_events[%d]: unrecognized event type %q", i, ae.Type)
		}
		result = append(result, codex_types.ToCodex([]types.AgentEvent{ae})...)
	}

	return result, nil
}

func isValidActionType(t types.ActionType) bool {
	switch t {
	case types.ActionThink, types.ActionToolCall, types.ActionMessage,
		types.ActionError, types.ActionDone, types.ActionStepStart, types.ActionStepFinish:
		return true
	}
	return false
}
