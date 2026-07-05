package run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteGrokConfigToml(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := writeGrokConfigToml(path, 8123); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		"models_base_url",
		"http://127.0.0.1:8123/v1",
		"mock-model",
		"api_backend = \"chat_completions\"",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("config.toml missing %q:\n%s", want, content)
		}
	}
}

func TestReadPort(t *testing.T) {
	port, err := readPort(strings.NewReader(":8080\n"))
	if err != nil {
		t.Fatal(err)
	}
	if port != 8080 {
		t.Fatalf("port = %d, want 8080", port)
	}
}

func TestResolveGrokHomeExplicit(t *testing.T) {
	explicit := filepath.Join(t.TempDir(), "explicit-home")
	t.Setenv("AGENT_RUNNER_CONFIG_HOME", "")
	t.Setenv("LLM_MOCK_GROK_HOME", explicit)

	home, cleanup, err := resolveGrokHome(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if home != explicit {
		t.Fatalf("home = %q, want %q", home, explicit)
	}
	if cleanup {
		t.Fatal("expected no cleanup for explicit home")
	}
}

func TestResolveGrokHomeAgentRunnerConfigHome(t *testing.T) {
	shared := filepath.Join(t.TempDir(), "shared-home")
	t.Setenv("AGENT_RUNNER_CONFIG_HOME", shared)
	t.Setenv("LLM_MOCK_GROK_HOME", "")

	home, cleanup, err := resolveGrokHome(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if home != shared {
		t.Fatalf("home = %q, want %q", home, shared)
	}
	if cleanup {
		t.Fatal("expected no cleanup for shared config home")
	}
}

func TestResolveGrokHomeDefaultTemp(t *testing.T) {
	t.Setenv("AGENT_RUNNER_CONFIG_HOME", "")
	t.Setenv("LLM_MOCK_GROK_HOME", "")
	tmp := t.TempDir()

	home, cleanup, err := resolveGrokHome(tmp, "")
	if err != nil {
		t.Fatal(err)
	}
	if !cleanup {
		t.Fatal("expected cleanup for default temp home")
	}
	if !strings.HasPrefix(home, tmp) {
		t.Fatalf("home %q not under temp dir %q", home, tmp)
	}
}