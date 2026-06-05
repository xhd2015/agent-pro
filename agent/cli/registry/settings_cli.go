package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type AgentRunnerID string

const (
	AgentRunnerCodex    AgentRunnerID = "codex"
	AgentRunnerOpencode AgentRunnerID = "opencode"
	AgentRunnerCursor   AgentRunnerID = "cursor"
	AgentRunnerFakeCodex AgentRunnerID = "fake-codex"
)

const (
	CursorCLIPathSettingKey        = "cursor_cli_path"
	CodexCLIPathSettingKey         = "codex_cli_path"
	OpencodeCLIPathSettingKey      = "opencode_cli_path"
	FakeCodexCLIPathSettingKey     = "fake_codex_cli_path"
	CodexAPIKeySettingKey          = "codex_api_key"
	AgentRunnerIDSettingKey        = "agent_runner_id"
	KBDefaultAgentRunnerIDSettingKey = "kb_default_agent_runner_id"
	ModelSettingKey                = "model"
	ModelsByAgentRunnerSettingKey  = "models_by_agent_runner"
)

type Settings struct {
	Model                  string            `json:"model,omitempty"`
	AgentID                string            `json:"agent_id,omitempty"`
	AgentRunnerID          string            `json:"agent_runner_id,omitempty"`
	KBDefaultAgentRunnerID string            `json:"kb_default_agent_runner_id,omitempty"`
	CursorCLIPath          string            `json:"cursor_cli_path,omitempty"`
	CodexCLIPath           string            `json:"codex_cli_path,omitempty"`
	OpencodeCLIPath        string            `json:"opencode_cli_path,omitempty"`
	FakeCodexCLIPath       string            `json:"fake_codex_cli_path,omitempty"`
	CodexAPIKey            string            `json:"codex_api_key,omitempty"`
	DisableSubAgents       bool              `json:"disable_sub_agents,omitempty"`
	ModelsByAgentRunner    map[string]string `json:"models_by_agent_runner,omitempty"`
}

const (
	EnvCursorCLIPath   = "AGENT_RUNNER_CURSOR_PATH"
	EnvCodexCLIPath    = "AGENT_RUNNER_CODEX_PATH"
	EnvOpencodeCLIPath = "AGENT_RUNNER_OPENCODE_PATH"
	EnvFakeCodexCLIPath = "AGENT_RUNNER_FAKE_CODEX_PATH"
)

func LoadEnvCLIPath(envKey string) string {
	return strings.TrimSpace(os.Getenv(envKey))
}

func ResolveConfiguredCLIPath(settingsPath, settingKey, envKey, defaultPath string, fallback func() (string, error)) (string, error) {
	if strings.TrimSpace(defaultPath) != "" {
		return strings.TrimSpace(defaultPath), nil
	}
	if envVal := LoadEnvCLIPath(envKey); envVal != "" {
		return envVal, nil
	}
	if configured := LoadConfiguredStringSetting(settingsPath, settingKey); configured != "" {
		return configured, nil
	}
	if fallback == nil {
		return "", fmt.Errorf("no CLI resolver configured for %s", settingKey)
	}
	return fallback()
}

func LoadConfiguredCLIPath(settingsPath string, settingKey string) string {
	return LoadConfiguredStringSetting(settingsPath, settingKey)
}

func LoadConfiguredStringSetting(settingsPath string, settingKey string) string {
	settings := LoadSettingsMap(settingsPath)
	switch settingKey {
	case CursorCLIPathSettingKey:
		return strings.TrimSpace(settings.CursorCLIPath)
	case CodexCLIPathSettingKey:
		return strings.TrimSpace(settings.CodexCLIPath)
	case OpencodeCLIPathSettingKey:
		return strings.TrimSpace(settings.OpencodeCLIPath)
	case FakeCodexCLIPathSettingKey:
		return strings.TrimSpace(settings.FakeCodexCLIPath)
	case CodexAPIKeySettingKey:
		return strings.TrimSpace(settings.CodexAPIKey)
	case AgentRunnerIDSettingKey:
		return strings.TrimSpace(settings.AgentRunnerID)
	case KBDefaultAgentRunnerIDSettingKey:
		return strings.TrimSpace(settings.KBDefaultAgentRunnerID)
	case ModelSettingKey:
		return strings.TrimSpace(settings.Model)
	default:
		return ""
	}
}

func LoadConfiguredStringMapSetting(settingsPath string, settingKey string) map[string]string {
	settings := LoadSettingsMap(settingsPath)
	if settingKey != ModelsByAgentRunnerSettingKey || len(settings.ModelsByAgentRunner) == 0 {
		return nil
	}
	out := make(map[string]string, len(settings.ModelsByAgentRunner))
	for runnerID, model := range settings.ModelsByAgentRunner {
		if trimmed := strings.TrimSpace(model); trimmed != "" {
			out[runnerID] = trimmed
		}
	}
	return out
}

func LoadSettingsMap(settingsPath string) Settings {
	if strings.TrimSpace(settingsPath) == "" {
		return Settings{}
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil || len(data) == 0 {
		return Settings{}
	}
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}
	}
	return settings
}
