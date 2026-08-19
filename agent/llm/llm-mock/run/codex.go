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

// EnvExtraMCPTOMLFile, when set, is a TOML file appended to the generated
// CODEX_HOME/config.toml (typically [mcp_servers.*] blocks for tests).
const EnvExtraMCPTOMLFile = "LLM_MOCK_EXTRA_MCP_TOML_FILE"

// RunCodexOptions configures optional behavior for RunCodex.
type RunCodexOptions struct {
	MockEventsPreset string // --mock-events-preset; passed to mock server
	MockEventsFile   string // --mock-events-file; AgentEvent JSONL for genQueue
	LogEventsPath    string // --log-events output path; passed to mock server as --agent-events-file
	LogHTTPPath      string // --log-http output path; passed to mock server as --log-http
}

// RunCodex starts the mock server in the background, configures an isolated CODEX_HOME,
// runs codex in the foreground, and tears down the mock when codex exits.
func RunCodex(codexArgs []string, opts RunCodexOptions) error {
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

	tmpDir, err := os.MkdirTemp("", "llm-mock-run-codex-*")
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

	mockCmd, port, err := startMockServer(ctx, mergedConfigPath, opts.MockEventsPreset, opts.MockEventsFile, opts.LogEventsPath, opts.LogHTTPPath)
	if err != nil {
		return err
	}
	defer func() {
		teardownStart := time.Now()
		runCodexDebugf("mock teardown: cancel context")
		cancel()
		if mockCmd.Process != nil {
			runCodexDebugf("mock teardown: SIGTERM pid=%d", mockCmd.Process.Pid)
			_ = mockCmd.Process.Signal(syscall.SIGTERM)
			waitStart := time.Now()
			waitErr := mockCmd.Wait()
			runCodexDebugf("mock teardown: Wait returned after %s err=%v", since(waitStart), waitErr)
		}
		runCodexDebugf("mock teardown done after %s", since(teardownStart))
	}()

	codexHome, err := resolveCodexHome(tmpDir)
	if err != nil {
		return err
	}

	if err := writeCodexConfigToml(filepath.Join(codexHome, "config.toml"), port); err != nil {
		return err
	}

	if err := writeCodexModelsCache(filepath.Join(codexHome, "models_cache.json")); err != nil {
		return err
	}

	codexEnv := append(os.Environ(),
		"CODEX_HOME="+codexHome,
		"OPENAI_API_KEY=sk-mock",
	)

	fmt.Fprintf(os.Stderr, "CODEX_HOME=%s\n", codexHome)

	codexCmd, usedHook, err := buildCodexCommand(codexArgs)
	if err != nil {
		return err
	}
	codexCmd.Env = codexEnv
	codexCmd.Stdin = os.Stdin
	codexCmd.Stdout = os.Stdout
	codexCmd.Stderr = os.Stderr

	runCodexDebugf("starting codex (hook=%v) args=%q codexHome=%q mockPort=%d",
		usedHook, codexArgs, codexHome, port)
	codexStart := time.Now()
	runErr := codexCmd.Run()
	codexElapsed := since(codexStart)
	if runErr != nil {
		runCodexDebugf("codex exited with error after %s: %v", codexElapsed, runErr)
	} else {
		runCodexDebugf("codex exited ok after %s", codexElapsed)
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("run codex: %w", runErr)
	}
	return nil
}

func resolveCodexHome(tmpDir string) (string, error) {
	if v := os.Getenv("LLM_MOCK_CODEX_HOME"); v != "" {
		if err := os.MkdirAll(v, 0755); err != nil {
			return "", fmt.Errorf("mkdir codex home: %w", err)
		}
		return v, nil
	}
	home := filepath.Join(tmpDir, "codex-home")
	if err := os.MkdirAll(home, 0755); err != nil {
		return "", fmt.Errorf("mkdir codex home: %w", err)
	}
	return home, nil
}

func buildCodexCommand(codexArgs []string) (*exec.Cmd, bool, error) {
	if hook := os.Getenv("LLM_MOCK_RUN_CODEX_COMMAND"); hook != "" {
		return exec.Command("sh", "-c", hook), true, nil
	}
	if len(codexArgs) == 0 {
		return exec.Command("codex"), false, nil
	}
	return exec.Command("codex", codexArgs...), false, nil
}

type codexModelsCache struct {
	FetchedAt     string            `json:"fetched_at"`
	Etag          string            `json:"etag"`
	ClientVersion string            `json:"client_version"`
	Models        []codexModelEntry `json:"models"`
}

type codexModelEntry struct {
	Slug                          string                `json:"slug"`
	DisplayName                   string                `json:"display_name"`
	Description                   string                `json:"description"`
	DefaultReasoningLevel         string                `json:"default_reasoning_level"`
	SupportedReasoningLevels      []codexReasoningLevel `json:"supported_reasoning_levels"`
	ShellType                     string                `json:"shell_type"`
	Visibility                    string                `json:"visibility"`
	SupportedInAPI                bool                  `json:"supported_in_api"`
	Priority                      int                   `json:"priority"`
	AdditionalSpeedTiers          []string              `json:"additional_speed_tiers"`
	ServiceTiers                  []codexServiceTier    `json:"service_tiers"`
	DefaultServiceTier            string                `json:"default_service_tier"`
	AvailabilityNUX               codexAvailabilityNUX  `json:"availability_nux"`
	Upgrade                       *string               `json:"upgrade"`
	BaseInstructions              string                `json:"base_instructions"`
	ModelMessages                 codexModelMessages    `json:"model_messages"`
	SupportsReasoningSummaries    bool                  `json:"supports_reasoning_summaries"`
	DefaultReasoningSummary       string                `json:"default_reasoning_summary"`
	SupportVerbosity              bool                  `json:"support_verbosity"`
	DefaultVerbosity              string                `json:"default_verbosity"`
	ApplyPatchToolType            string                `json:"apply_patch_tool_type"`
	WebSearchToolType             string                `json:"web_search_tool_type"`
	TruncationPolicy              codexTruncationPolicy `json:"truncation_policy"`
	SupportsParallelToolCalls     bool                  `json:"supports_parallel_tool_calls"`
	SupportsImageDetailOriginal   bool                  `json:"supports_image_detail_original"`
	ContextWindow                 int                   `json:"context_window"`
	MaxContextWindow              int                   `json:"max_context_window"`
	CompHash                      string                `json:"comp_hash"`
	EffectiveContextWindowPercent int                   `json:"effective_context_window_percent"`
	ExperimentalSupportedTools    []string              `json:"experimental_supported_tools"`
	InputModalities               []string              `json:"input_modalities"`
	SupportsSearchTool            bool                  `json:"supports_search_tool"`
	UseResponsesLite              bool                  `json:"use_responses_lite"`
}

type codexReasoningLevel struct {
	Effort      string `json:"effort"`
	Description string `json:"description"`
}

type codexServiceTier struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type codexAvailabilityNUX struct {
	Message string `json:"message"`
}

type codexModelMessages struct {
	InstructionsTemplate  string            `json:"instructions_template"`
	InstructionsVariables map[string]string `json:"instructions_variables"`
}

type codexTruncationPolicy struct {
	Mode  string `json:"mode"`
	Limit int    `json:"limit"`
}

func resolveCodexClientVersion() string {
	out, err := exec.Command("codex", "--version").Output()
	if err != nil {
		return "llm-mock"
	}
	line := strings.TrimSpace(string(out))
	if i := strings.LastIndex(line, " "); i >= 0 && i < len(line)-1 {
		return strings.TrimSpace(line[i+1:])
	}
	return "llm-mock"
}

func writeCodexModelsCache(path string) error {
	mockInstructions := "You are a mock model for local LLM testing."
	cache := codexModelsCache{
		FetchedAt:     time.Now().UTC().Format(time.RFC3339),
		Etag:          "llm-mock",
		ClientVersion: resolveCodexClientVersion(),
		Models: []codexModelEntry{
			{
				Slug:                  "mock-model",
				DisplayName:           "Mock Model",
				Description:           "LLM mock model for local testing",
				DefaultReasoningLevel: "medium",
				SupportedReasoningLevels: []codexReasoningLevel{
					{Effort: "low", Description: "Fast responses with lighter reasoning"},
					{Effort: "medium", Description: "Balances speed and reasoning depth for everyday tasks"},
					{Effort: "high", Description: "Greater reasoning depth for complex problems"},
				},
				ShellType:            "shell_command",
				Visibility:           "list",
				SupportedInAPI:       true,
				Priority:             0,
				AdditionalSpeedTiers: []string{"fast"},
				ServiceTiers: []codexServiceTier{
					{ID: "priority", Name: "Fast", Description: "1.5x speed, increased usage"},
				},
				DefaultServiceTier: "priority",
				AvailabilityNUX:    codexAvailabilityNUX{Message: ""},
				Upgrade:            nil,
				BaseInstructions:   mockInstructions,
				ModelMessages: codexModelMessages{
					InstructionsTemplate:  mockInstructions,
					InstructionsVariables: map[string]string{},
				},
				SupportsReasoningSummaries:    true,
				DefaultReasoningSummary:       "none",
				SupportVerbosity:              true,
				DefaultVerbosity:              "low",
				ApplyPatchToolType:            "freeform",
				WebSearchToolType:             "text_and_image",
				TruncationPolicy:              codexTruncationPolicy{Mode: "tokens", Limit: 10000},
				SupportsParallelToolCalls:     true,
				SupportsImageDetailOriginal:   true,
				ContextWindow:                 272000,
				MaxContextWindow:              272000,
				CompHash:                      "llm-mock",
				EffectiveContextWindowPercent: 95,
				ExperimentalSupportedTools:    []string{},
				InputModalities:               []string{"text", "image"},
				SupportsSearchTool:            true,
				UseResponsesLite:              false,
			},
		},
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal codex models cache: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write codex models_cache.json: %w", err)
	}
	return nil
}

func writeCodexConfigToml(path string, port int) error {
	baseURL := fmt.Sprintf("http://127.0.0.1:%d/v1", port)
	content := fmt.Sprintf(`model = "mock-model"
model_provider = "llm-mock"
approval_policy = "never"

[model_providers.llm-mock]
name = "LLM Mock"
base_url = %q
wire_api = "responses"
env_key = "OPENAI_API_KEY"

[features]
shell_tool = true
unified_exec = true
`, baseURL)
	if extraFile := strings.TrimSpace(os.Getenv(EnvExtraMCPTOMLFile)); extraFile != "" {
		b, err := os.ReadFile(extraFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", EnvExtraMCPTOMLFile, err)
		}
		content += "\n" + string(b)
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write codex config.toml: %w", err)
	}
	return nil
}
