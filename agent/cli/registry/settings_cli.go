package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	CursorCLIPathSettingKey       = "cursor_cli_path"
	CodexCLIPathSettingKey        = "codex_cli_path"
	OpencodeCLIPathSettingKey     = "opencode_cli_path"
	CodexAPIKeySettingKey         = "codex_api_key"
	ProviderIDSettingKey          = "provider_id"
	KBDefaultProviderIDSettingKey = "kb_default_provider_id"
	ModelSettingKey               = "model"
	ModelsByProviderSettingKey    = "models_by_provider"
)

type Settings struct {
	Model               string            `json:"model,omitempty"`
	AgentID             string            `json:"agent_id,omitempty"`
	ProviderID          string            `json:"provider_id,omitempty"`
	KBDefaultProviderID string            `json:"kb_default_provider_id,omitempty"`
	CursorCLIPath       string            `json:"cursor_cli_path,omitempty"`
	CodexCLIPath        string            `json:"codex_cli_path,omitempty"`
	OpencodeCLIPath     string            `json:"opencode_cli_path,omitempty"`
	CodexAPIKey         string            `json:"codex_api_key,omitempty"`
	DisableSubAgents    bool              `json:"disable_sub_agents,omitempty"`
	ModelsByProvider    map[string]string `json:"models_by_provider,omitempty"`
}

func ResolveConfiguredCLIPath(settingsPath string, settingKey string, defaultPath string, fallback func() (string, error)) (string, error) {
	if configured := LoadConfiguredStringSetting(settingsPath, settingKey); configured != "" {
		return configured, nil
	}
	if strings.TrimSpace(defaultPath) != "" {
		return strings.TrimSpace(defaultPath), nil
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
	case CodexAPIKeySettingKey:
		return strings.TrimSpace(settings.CodexAPIKey)
	case ProviderIDSettingKey:
		return strings.TrimSpace(settings.ProviderID)
	case KBDefaultProviderIDSettingKey:
		return strings.TrimSpace(settings.KBDefaultProviderID)
	case ModelSettingKey:
		return strings.TrimSpace(settings.Model)
	default:
		return ""
	}
}

func LoadConfiguredStringMapSetting(settingsPath string, settingKey string) map[string]string {
	settings := LoadSettingsMap(settingsPath)
	if settingKey != ModelsByProviderSettingKey || len(settings.ModelsByProvider) == 0 {
		return nil
	}
	out := make(map[string]string, len(settings.ModelsByProvider))
	for providerID, model := range settings.ModelsByProvider {
		if trimmed := strings.TrimSpace(model); trimmed != "" {
			out[providerID] = trimmed
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
