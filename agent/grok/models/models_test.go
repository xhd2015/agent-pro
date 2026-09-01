package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListMergesConfigAndCache(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	cfg := `
[models]
default = "grok-4.5"

[model."grok-4.5"]
context_window = 300000

[model."compass-local-switch-glm-5-2"]
name = "AIS - GLM-5.2"
`
	if err := os.WriteFile(filepath.Join(home, DefaultConfigFile), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := `{
  "models": {
    "grok-4.6": {
      "info": { "name": "Grok 4.6" }
    },
    "grok-4.5": {
      "info": { "name": "Grok 4.5" }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(home, ModelsCacheFile), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}

	cat, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Default != "grok-4.5" {
		t.Fatalf("Default=%q", cat.Default)
	}
	if !cat.FromConfig || !cat.FromCache {
		t.Fatalf("FromConfig=%v FromCache=%v", cat.FromConfig, cat.FromCache)
	}
	want := []Model{
		{ID: "compass-local-switch-glm-5-2", Source: DefaultConfigFile, DisplayName: "AIS - GLM-5.2"},
		{ID: "grok-4.5", Source: DefaultConfigFile, DisplayName: "Grok 4.5"},
		{ID: "grok-4.6", Source: ModelsCacheFile, DisplayName: "Grok 4.6"},
	}
	if len(cat.Models) != len(want) {
		t.Fatalf("Models=%+v want %+v", cat.Models, want)
	}
	for i := range want {
		if cat.Models[i] != want[i] {
			t.Fatalf("Models[%d]=%+v want %+v", i, cat.Models[i], want[i])
		}
	}
}

func TestListPreferConfigSourceFillDisplayFromCache(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	cfg := `
[models]
default = "grok-4.5"

[model."grok-4.5"]
context_window = 1
`
	if err := os.WriteFile(filepath.Join(home, DefaultConfigFile), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := `{"models":{"grok-4.5":{"info":{"name":"Grok 4.5"}}}}`
	if err := os.WriteFile(filepath.Join(home, ModelsCacheFile), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}
	cat, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Models) != 1 {
		t.Fatalf("Models=%+v", cat.Models)
	}
	m := cat.Models[0]
	if m.ID != "grok-4.5" || m.Source != DefaultConfigFile || m.DisplayName != "Grok 4.5" {
		t.Fatalf("model=%+v", m)
	}
}

func TestListMissingHomeEmpty(t *testing.T) {
	t.Parallel()
	home := filepath.Join(t.TempDir(), "missing-grok-home")
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

func TestListCacheOnly(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	cache := `{"models":{"only-cache":{"info":{"name":"Only Cache"}}}}`
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
	if len(cat.Models) != 1 {
		t.Fatalf("Models=%+v", cat.Models)
	}
	m := cat.Models[0]
	if m.ID != "only-cache" || m.Source != ModelsCacheFile || m.DisplayName != "Only Cache" {
		t.Fatalf("model=%+v", m)
	}
}
