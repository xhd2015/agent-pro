package slackutil

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SlackConfig is the shared Slack JSON configuration shape.
type SlackConfig struct {
	Source             string            `json:"source"`
	BotToken           string            `json:"botToken"`
	AppToken           string            `json:"appToken"`
	DefaultChannelId   string            `json:"defaultChannelId"`
	DefaultChannelName string            `json:"defaultChannelName"`
	KnownChannels      map[string]string `json:"knownChannels"`
	Config             json.RawMessage   `json:"config"`
}

// Load reads and parses a Slack JSON config file.
func Load(path string) (*SlackConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg SlackConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ConfigDisplayPath returns an absolute path for startup logging, falling back to
// the original path when Abs fails.
func ConfigDisplayPath(path string) string {
	if path == "" {
		return "(none)"
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		return abs
	}
	return path
}