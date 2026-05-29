package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tidwall/jsonc"
)

type Data = map[string]interface{}

type Config struct {
	Path string
	Data Data
}

func Read(dir string) (*Config, error) {
	opencodeDir := filepath.Join(dir, ".opencode")

	var configPath string
	var raw []byte

	for _, name := range []string{"opencode.jsonc", "opencode.json"} {
		p := filepath.Join(opencodeDir, name)
		data, err := os.ReadFile(p)
		if err == nil {
			configPath = p
			raw = data
			break
		}
	}

	if configPath == "" {
		configPath = filepath.Join(opencodeDir, "opencode.json")
		return &Config{Path: configPath, Data: Data{}}, nil
	}

	cleaned := jsonc.ToJSON(raw)

	var data Data
	if err := json.Unmarshal(cleaned, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w\n\nRaw content:\n%s", configPath, err, string(raw))
	}

	return &Config{Path: configPath, Data: data}, nil
}

func (c *Config) Write() error {
	output, err := json.MarshalIndent(c.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	output = append(output, '\n')

	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".opencode-config-*.json")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(output); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := os.Rename(tmpPath, c.Path); err != nil {
		return fmt.Errorf("replace %s: %w", c.Path, err)
	}
	return nil
}

func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
