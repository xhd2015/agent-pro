// Package models lists available Grok CLI models from on-disk files only
// (~/.grok/config.toml and models_cache.json). It does not spawn `grok models`.
package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultConfigFile is the Grok home config name.
	DefaultConfigFile = "config.toml"
	// ModelsCacheFile is the Grok home models cache name.
	ModelsCacheFile = "models_cache.json"
)

// Catalog is the merged file-backed model list for a Grok home.
type Catalog struct {
	// Home is the resolved Grok data directory used for the listing.
	Home string `json:"home"`
	// Default is [models].default from config.toml (empty when unset/missing).
	Default string `json:"default,omitempty"`
	// Models is the sorted unique union of config.toml [model."…"] keys and
	// models_cache.json model ids. Empty when neither source yields ids.
	Models []string `json:"models"`
	// FromConfig is true when config.toml was read successfully.
	FromConfig bool `json:"from_config"`
	// FromCache is true when models_cache.json was read successfully.
	FromCache bool `json:"from_cache"`
}

type configFile struct {
	Models configModelsSection `toml:"models"`
	Model  map[string]any      `toml:"model"`
}

type configModelsSection struct {
	Default                 string `toml:"default"`
	DefaultReasoningEffort  string `toml:"default_reasoning_effort"`
}

type modelsCacheFile struct {
	Models map[string]json.RawMessage `json:"models"`
}

// DefaultHome returns $GROK_HOME when set, otherwise $HOME/.grok.
// Empty HOME falls back to a temp-dir based path (same idea as agenttty.GrokHome).
func DefaultHome() string {
	if v := strings.TrimSpace(os.Getenv("GROK_HOME")); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return filepath.Join(os.TempDir(), ".grok")
	}
	return filepath.Join(home, ".grok")
}

// List returns the file-backed catalog for home.
// Empty home uses DefaultHome(). Missing files are soft: empty Models, no error
// unless a present file is unreadable/invalid.
func List(home string) (Catalog, error) {
	home = strings.TrimSpace(home)
	if home == "" {
		home = DefaultHome()
	}
	cat := Catalog{Home: home, Models: nil}

	seen := map[string]struct{}{}
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		cat.Models = append(cat.Models, id)
	}

	cfgPath := filepath.Join(home, DefaultConfigFile)
	if raw, err := os.ReadFile(cfgPath); err == nil {
		var cfg configFile
		if _, err := toml.Decode(string(raw), &cfg); err != nil {
			return Catalog{}, fmt.Errorf("grok models: parse %s: %w", cfgPath, err)
		}
		cat.FromConfig = true
		cat.Default = strings.TrimSpace(cfg.Models.Default)
		for id := range cfg.Model {
			add(id)
		}
	} else if !os.IsNotExist(err) {
		return Catalog{}, fmt.Errorf("grok models: read %s: %w", cfgPath, err)
	}

	cachePath := filepath.Join(home, ModelsCacheFile)
	if raw, err := os.ReadFile(cachePath); err == nil {
		var cache modelsCacheFile
		if err := json.Unmarshal(raw, &cache); err != nil {
			return Catalog{}, fmt.Errorf("grok models: parse %s: %w", cachePath, err)
		}
		cat.FromCache = true
		for id := range cache.Models {
			add(id)
		}
	} else if !os.IsNotExist(err) {
		return Catalog{}, fmt.Errorf("grok models: read %s: %w", cachePath, err)
	}

	sort.Strings(cat.Models)
	if cat.Models == nil {
		cat.Models = []string{}
	}
	return cat, nil
}

// ListIDs is List(home).Models.
func ListIDs(home string) ([]string, error) {
	cat, err := List(home)
	if err != nil {
		return nil, err
	}
	return append([]string(nil), cat.Models...), nil
}
