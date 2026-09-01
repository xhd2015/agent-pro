package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListMergesConfigAndCache(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	config := `
model = "gpt-5.5"
model_reasoning_effort = "medium"
`
	if err := os.WriteFile(filepath.Join(home, DefaultConfigFile), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := `{
  "models": [
    {
      "slug": "gpt-5.6-sol",
      "display_name": "GPT-5.6-Sol",
      "default_reasoning_level": "medium",
      "visibility": "list",
      "supported_reasoning_levels": [
        {"effort": "low", "description": "a"},
        {"effort": "ultra", "description": "b"}
      ]
    },
    {"slug": "gpt-reserve", "visibility": "hide"},
    {"slug": "gpt-5.5", "visibility": "list"}
  ]
}`
	if err := os.WriteFile(filepath.Join(home, ModelsCacheFile), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}

	cat, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Default != "gpt-5.5" {
		t.Fatalf("Default=%q", cat.Default)
	}
	if !cat.FromConfig || !cat.FromCache {
		t.Fatalf("FromConfig=%v FromCache=%v", cat.FromConfig, cat.FromCache)
	}
	if len(cat.Models) != 2 {
		t.Fatalf("Models=%+v", cat.Models)
	}
	if cat.Models[0].ID != "gpt-5.6-sol" || cat.Models[0].Source != ModelsCacheFile || cat.Models[0].DisplayName != "GPT-5.6-Sol" {
		t.Fatalf("Models[0]=%+v", cat.Models[0])
	}
	if got := cat.Models[0].DefaultReasoning; got != "medium" {
		t.Fatalf("DefaultReasoning=%q", got)
	}
	if got := strings.Join(cat.Models[0].Reasoning, ","); got != "low,ultra" {
		t.Fatalf("Reasoning=%v", cat.Models[0].Reasoning)
	}
	if cat.Models[1].ID != "gpt-5.5" || cat.Models[1].Source != ModelsCacheFile {
		t.Fatalf("Models[1]=%+v", cat.Models[1])
	}
}

func TestListConfigModelNotInCachePrepended(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, DefaultConfigFile), []byte(`model = "openrouter/deepseek-chat"`), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := `{"models":[{"slug":"gpt-5.5","visibility":"list"}]}`
	if err := os.WriteFile(filepath.Join(home, ModelsCacheFile), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}

	cat, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Default != "openrouter/deepseek-chat" {
		t.Fatalf("Default=%q", cat.Default)
	}
	if len(cat.Models) != 2 {
		t.Fatalf("Models=%+v", cat.Models)
	}
	if cat.Models[0].ID != "openrouter/deepseek-chat" || cat.Models[0].Source != DefaultConfigFile {
		t.Fatalf("Models[0]=%+v", cat.Models[0])
	}
	if cat.Models[1].ID != "gpt-5.5" || cat.Models[1].Source != ModelsCacheFile {
		t.Fatalf("Models[1]=%+v", cat.Models[1])
	}
	if cat.Models[0].Reasoning != nil || cat.Models[0].DisplayName != "" {
		t.Fatalf("Models[0]=%+v", cat.Models[0])
	}
}

func TestListMissingHomeEmpty(t *testing.T) {
	t.Parallel()
	home := filepath.Join(t.TempDir(), "missing-codex-home")
	cat, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Default != "" || len(cat.Models) != 0 {
		t.Fatalf("cat=%+v", cat)
	}
	if cat.FromConfig || cat.FromCache {
		t.Fatalf("flags FromConfig=%v FromCache=%v", cat.FromConfig, cat.FromCache)
	}
}

func TestListInvalidConfigErrors(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, DefaultConfigFile), []byte("[[[not toml"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := List(home); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestListInvalidCacheErrors(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ModelsCacheFile), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := List(home); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestListCacheOnly(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	cache := `{"models":[{"slug":"gpt-5.6-luna","visibility":"list"}]}`
	if err := os.WriteFile(filepath.Join(home, ModelsCacheFile), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}
	cat, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if !cat.FromCache || cat.FromConfig {
		t.Fatalf("FromCache=%v FromConfig=%v", cat.FromCache, cat.FromConfig)
	}
	if len(cat.Models) != 1 || cat.Models[0].ID != "gpt-5.6-luna" || cat.Models[0].Source != ModelsCacheFile {
		t.Fatalf("Models=%+v", cat.Models)
	}
}

func TestDefaultHomeEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CODEX_HOME", tmp)
	if got := DefaultHome(); got != tmp {
		t.Fatalf("DefaultHome=%q want %q", got, tmp)
	}
}
