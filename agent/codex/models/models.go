// Package models lists available Codex CLI models from on-disk files only
// (~/.codex/models_cache.json and config.toml). It does not spawn `codex`.
package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultConfigFile is the Codex home config name.
	DefaultConfigFile = "config.toml"
	// ModelsCacheFile is the Codex home models cache name.
	ModelsCacheFile = "models_cache.json"
	// VisibilityList marks cache entries Codex shows in its model picker.
	VisibilityList = "list"
)

// Catalog is the merged file-backed model list for a Codex home.
type Catalog struct {
	// Home is the resolved Codex data directory used for the listing.
	Home string `json:"home"`
	// Default is the top-level `model` from config.toml (empty when unset/missing).
	Default string `json:"default,omitempty"`
	// Models is the picker model list: cache entries with visibility "list",
	// plus the configured model when the cache omits it, in cache order.
	Models []Model `json:"models"`
	// FromConfig is true when config.toml was read successfully.
	FromConfig bool `json:"from_config"`
	// FromCache is true when models_cache.json was read successfully.
	FromCache bool `json:"from_cache"`
}

// Model is one selectable Codex model in the unified CLI JSON shape.
type Model struct {
	// ID is the model id passed to codex (e.g. "gpt-5.6-sol"); for a
	// configured provider-qualified id it may look like "provider/model".
	ID string `json:"id"`
	// Source is the basename of the file that primarily contributed this id
	// (models_cache.json for cache rows; config.toml for config-only unions).
	Source string `json:"source"`
	// DisplayName is Codex's human name (cache only; empty otherwise).
	DisplayName string `json:"display_name,omitempty"`
	// DefaultReasoning is the cache's default_reasoning_level (empty when unknown).
	DefaultReasoning string `json:"default_reasoning,omitempty"`
	// Reasoning is the supported effort levels in cache order (low…ultra).
	Reasoning []string `json:"reasoning,omitempty"`
}

type configFile struct {
	Model string `toml:"model"`
}

type modelsCacheFile struct {
	Models []cacheModel `json:"models"`
}

type cacheModel struct {
	Slug                    string           `json:"slug"`
	DisplayName             string           `json:"display_name"`
	DefaultReasoning        string           `json:"default_reasoning_level"`
	Visibility              string           `json:"visibility"`
	SupportedReasoningLevel []reasoningLevel `json:"supported_reasoning_levels"`
}

type reasoningLevel struct {
	Effort string `json:"effort"`
}

// DefaultHome returns $CODEX_HOME when set, otherwise $HOME/.codex.
// Empty HOME falls back to a temp-dir based path (same idea as sessions home).
func DefaultHome() string {
	if v := strings.TrimSpace(os.Getenv("CODEX_HOME")); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return filepath.Join(os.TempDir(), ".codex")
	}
	return filepath.Join(home, ".codex")
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

	cfgPath := filepath.Join(home, DefaultConfigFile)
	if raw, err := os.ReadFile(cfgPath); err == nil {
		var cfg configFile
		if _, err := toml.Decode(string(raw), &cfg); err != nil {
			return Catalog{}, fmt.Errorf("codex models: parse %s: %w", cfgPath, err)
		}
		cat.FromConfig = true
		cat.Default = strings.TrimSpace(cfg.Model)
	} else if !os.IsNotExist(err) {
		return Catalog{}, fmt.Errorf("codex models: read %s: %w", cfgPath, err)
	}

	cachePath := filepath.Join(home, ModelsCacheFile)
	if raw, err := os.ReadFile(cachePath); err == nil {
		var cache modelsCacheFile
		if err := json.Unmarshal(raw, &cache); err != nil {
			return Catalog{}, fmt.Errorf("codex models: parse %s: %w", cachePath, err)
		}
		cat.FromCache = true
		for _, cm := range cache.Models {
			id := strings.TrimSpace(cm.Slug)
			if id == "" || strings.TrimSpace(cm.Visibility) != VisibilityList {
				continue
			}
			cat.Models = append(cat.Models, Model{
				ID:               id,
				Source:           ModelsCacheFile,
				DisplayName:      strings.TrimSpace(cm.DisplayName),
				DefaultReasoning: strings.TrimSpace(cm.DefaultReasoning),
				Reasoning:        efforts(cm.SupportedReasoningLevel),
			})
		}
	} else if !os.IsNotExist(err) {
		return Catalog{}, fmt.Errorf("codex models: read %s: %w", cachePath, err)
	}

	// Keep the configured model selectable even when the cache omits it
	// (e.g. a provider-qualified model from [model_providers.<name>]).
	if cat.Default != "" && !hasID(cat.Models, cat.Default) {
		cat.Models = append([]Model{{
			ID:     cat.Default,
			Source: DefaultConfigFile,
		}}, cat.Models...)
	}

	if cat.Models == nil {
		cat.Models = []Model{}
	}
	return cat, nil
}

// Slugs returns the model ids from List(home) (legacy name; values are Model.ID).
func Slugs(home string) ([]string, error) {
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

func efforts(levels []reasoningLevel) []string {
	out := make([]string, 0, len(levels))
	for _, l := range levels {
		if e := strings.TrimSpace(l.Effort); e != "" {
			out = append(out, e)
		}
	}
	return out
}

func hasID(models []Model, id string) bool {
	for _, m := range models {
		if m.ID == id {
			return true
		}
	}
	return false
}
