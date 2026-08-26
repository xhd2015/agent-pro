package explain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configFileName = "config.json"

// Config holds persisted explain preferences under the explain root.
type Config struct {
	Version     int    `json:"version"`
	AgentRunner string `json:"agent_runner,omitempty"`
	Model       string `json:"model,omitempty"`
}

// explainRoot returns the explain data root (parent of sessions/).
// Honors AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME when set.
func explainRoot() (string, error) {
	if debugHome := strings.TrimSpace(os.Getenv(debugConfigHomeEnv)); debugHome != "" {
		return filepath.Clean(debugHome), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, defaultSessionsBaseDir), nil
}

func configPath() (string, error) {
	root, err := explainRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, configFileName), nil
}

// loadConfig reads config.json. Missing file returns (empty Config, nil).
// Corrupt JSON returns an error (fail hard).
func loadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config.json: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return Config{}, nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config.json: %w", err)
	}
	return cfg, nil
}

// loadConfigMap reads config.json as a generic map so unknown keys are preserved.
// Missing file returns (nil, nil).
func loadConfigMap() (map[string]interface{}, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config.json: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]interface{}{}, nil
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse config.json: %w", err)
	}
	if root == nil {
		root = map[string]interface{}{}
	}
	return root, nil
}

func saveConfigMap(root map[string]interface{}) error {
	if root == nil {
		root = map[string]interface{}{}
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config.json: %w", err)
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write config.json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write config.json: %w", err)
	}
	return nil
}

type setConfigPrefs struct {
	agentRunner      string
	model            string
	clearAgentRunner bool
	clearModel       bool
}

func (p setConfigPrefs) any() bool {
	return strings.TrimSpace(p.agentRunner) != "" ||
		strings.TrimSpace(p.model) != "" ||
		p.clearAgentRunner ||
		p.clearModel
}

func runShowConfig() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("{}")
			return nil
		}
		return fmt.Errorf("read config.json: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println("{}")
		return nil
	}
	var any interface{}
	if err := json.Unmarshal(data, &any); err != nil {
		return fmt.Errorf("parse config.json: %w", err)
	}
	out, err := json.MarshalIndent(any, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config.json: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

func runSetConfig(prefs setConfigPrefs) error {
	if !prefs.any() {
		return fmt.Errorf("--set-config requires --agent-runner/--model and/or --clear-agent-runner/--clear-model")
	}
	agentRunner := strings.TrimSpace(prefs.agentRunner)
	model := strings.TrimSpace(prefs.model)
	if agentRunner != "" {
		if !supportedAgentRunners[agentRunner] {
			return fmt.Errorf("unsupported agent runner: %s (supported: opencode, codex, grok, commandcode)", agentRunner)
		}
	}
	if prefs.clearAgentRunner && agentRunner != "" {
		return fmt.Errorf("--clear-agent-runner is mutually exclusive with --agent-runner")
	}
	if prefs.clearModel && model != "" {
		return fmt.Errorf("--clear-model is mutually exclusive with --model")
	}

	root, err := loadConfigMap()
	if err != nil {
		return err
	}
	if root == nil {
		root = map[string]interface{}{}
	}
	if _, ok := root["version"]; !ok {
		root["version"] = 1
	}

	if prefs.clearAgentRunner {
		delete(root, "agent_runner")
	}
	if agentRunner != "" {
		root["agent_runner"] = agentRunner
	}
	if prefs.clearModel {
		delete(root, "model")
	}
	if model != "" {
		root["model"] = model
	}

	return saveConfigMap(root)
}
