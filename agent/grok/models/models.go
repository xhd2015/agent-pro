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
	Models []Model `json:"models"`
	// FromConfig is true when config.toml was read successfully.
	FromConfig bool `json:"from_config"`
	// FromCache is true when models_cache.json was read successfully.
	FromCache bool `json:"from_cache"`
}

// Model is one selectable Grok model in the unified CLI JSON shape.
type Model struct {
	// ID is the model id passed to grok.
	ID string `json:"id"`
	// Source is the basename of the file that primarily contributed this id
	// (config.toml preferred over models_cache.json when both list it).
	Source string `json:"source"`
	// DisplayName is config `name` when set, else cache info.name.
	DisplayName string `json:"display_name,omitempty"`
}

type configFile struct {
	Models configModelsSection `toml:"models"`
	Model  map[string]any      `toml:"model"`
}

type configModelsSection struct {
	Default                string `toml:"default"`
	DefaultReasoningEffort string `toml:"default_reasoning_effort"`
}

type modelsCacheFile struct {
	Models map[string]json.RawMessage `json:"models"`
}

type cacheModelEntry struct {
	Info cacheModelInfo `json:"info"`
}

type cacheModelInfo struct {
	Name string `json:"name"`
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
	byID := map[string]Model{}

	cfgPath := filepath.Join(home, DefaultConfigFile)
	if raw, err := os.ReadFile(cfgPath); err == nil {
		var cfg configFile
		if _, err := toml.Decode(string(raw), &cfg); err != nil {
			return Catalog{}, fmt.Errorf("grok models: parse %s: %w", cfgPath, err)
		}
		cat.FromConfig = true
		cat.Default = strings.TrimSpace(cfg.Models.Default)
		for id, v := range cfg.Model {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			byID[id] = Model{
				ID:          id,
				Source:      DefaultConfigFile,
				DisplayName: configModelName(v),
			}
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
		for id, entryRaw := range cache.Models {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			cacheName := cacheModelName(entryRaw)
			if existing, ok := byID[id]; ok {
				// Prefer config for source; fill display name from cache when config omitted it.
				if existing.DisplayName == "" && cacheName != "" {
					existing.DisplayName = cacheName
					byID[id] = existing
				}
				continue
			}
			byID[id] = Model{
				ID:          id,
				Source:      ModelsCacheFile,
				DisplayName: cacheName,
			}
		}
	} else if !os.IsNotExist(err) {
		return Catalog{}, fmt.Errorf("grok models: read %s: %w", cachePath, err)
	}

	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	cat.Models = make([]Model, 0, len(ids))
	for _, id := range ids {
		cat.Models = append(cat.Models, byID[id])
	}
	return cat, nil
}

// ListIDs returns List(home) model ids in catalog order.
func ListIDs(home string) ([]string, error) {
	cat, err := List(home)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(cat.Models))
	for _, m := range cat.Models {
		out = append(out, m.ID)
	}
	return out, nil
}

func configModelName(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	name, _ := m["name"].(string)
	return strings.TrimSpace(name)
}

func cacheModelName(raw json.RawMessage) string {
	var entry cacheModelEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return ""
	}
	return strings.TrimSpace(entry.Info.Name)
}
