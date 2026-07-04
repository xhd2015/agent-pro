package mockconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPathPriority(t *testing.T) {
	t.Setenv("LLM_MOCK_CONFIG_FILE", "/file.json")
	t.Setenv("LLM_MOCK_CONFIG", "/legacy.json")

	got, err := ResolveConfigPath("/flag.json")
	if err != nil {
		t.Fatalf("ResolveConfigPath(flag): %v", err)
	}
	if got != "/flag.json" {
		t.Fatalf("flag priority: got %q", got)
	}

	got, err = ResolveConfigPath("")
	if err != nil {
		t.Fatalf("ResolveConfigPath(file env): %v", err)
	}
	if got != "/file.json" {
		t.Fatalf("file env priority: got %q", got)
	}

	t.Setenv("LLM_MOCK_CONFIG_FILE", "")
	got, err = ResolveConfigPath("")
	if err != nil {
		t.Fatalf("ResolveConfigPath(legacy env): %v", err)
	}
	if got != "/legacy.json" {
		t.Fatalf("legacy env priority: got %q", got)
	}

	t.Setenv("LLM_MOCK_CONFIG", "")
	got, err = ResolveConfigPath("")
	if err != nil {
		t.Fatalf("ResolveConfigPath(none): %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty path for default config, got %q", got)
	}
}

func TestLoadMergedDefaultEmptyConfig(t *testing.T) {
	t.Setenv("LLM_MOCK_CONFIG_FILE", "")
	t.Setenv("LLM_MOCK_CONFIG", "")
	t.Setenv("LLM_MOCK_EVENTS_FILE", "")

	loaded, err := LoadMerged("")
	if err != nil {
		t.Fatalf("LoadMerged: %v", err)
	}
	if len(loaded.Exchanges) != 0 {
		t.Fatalf("expected 0 exchanges, got %d", len(loaded.Exchanges))
	}
	if loaded.Config.Port != 8080 {
		t.Fatalf("default port = %d, want 8080", loaded.Config.Port)
	}
}

func TestParseAndMergeAppendsEvents(t *testing.T) {
	configJSON := []byte(`{
  "port": 9090,
  "exchanges": [
    {
      "request": {"role": "user", "content": "first", "index": -1},
      "response": {"content": "from-config", "finish_reason": "stop"}
    }
  ]
}`)
	eventsPath := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(eventsPath, []byte(
		`{"request":{"role":"user","content":"second","index":-1},"response":{"content":"from-events","finish_reason":"stop"}}`+"\n",
	), 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := ParseAndMerge(configJSON, eventsPath)
	if err != nil {
		t.Fatalf("ParseAndMerge: %v", err)
	}
	if len(loaded.Exchanges) != 2 {
		t.Fatalf("expected 2 exchanges, got %d", len(loaded.Exchanges))
	}
	if loaded.Exchanges[1].Exchange.Response.Content == nil || *loaded.Exchanges[1].Exchange.Response.Content != "from-events" {
		t.Fatalf("second exchange not from events file")
	}
	if loaded.EffectiveIndices[0] != 0 || loaded.EffectiveIndices[1] != 1 {
		t.Fatalf("effective indices = %v, want [0 1]", loaded.EffectiveIndices)
	}
}

func TestParseAndMergeDuplicateExplicitIndex(t *testing.T) {
	configJSON := []byte(`{
  "port": 8080,
  "exchanges": [
    {
      "request": {"role": "user", "content": "first", "index": 0},
      "response": {"content": "one", "finish_reason": "stop"}
    }
  ]
}`)
	eventsPath := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(eventsPath, []byte(
		`{"request":{"role":"user","content":"second","index":0},"response":{"content":"two","finish_reason":"stop"}}`+"\n",
	), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAndMerge(configJSON, eventsPath); err == nil {
		t.Fatal("expected duplicate index error")
	}
}