package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile_Basic(t *testing.T) {
	content := `
model = "gpt-5.5"
model_reasoning_effort = "medium"
default_permissions = ":danger-full-access"
`

	cfg := testParse(t, content)
	if cfg.Model != "gpt-5.5" {
		t.Errorf("Model = %q, want %q", cfg.Model, "gpt-5.5")
	}
	if cfg.ModelReasoningEffort != "medium" {
		t.Errorf("ModelReasoningEffort = %q, want %q", cfg.ModelReasoningEffort, "medium")
	}
	if cfg.DefaultPermissions != ":danger-full-access" {
		t.Errorf("DefaultPermissions = %q, want %q", cfg.DefaultPermissions, ":danger-full-access")
	}
}

func TestReadFile_ModelProviders(t *testing.T) {
	content := `
[model_providers.llm-proxy]
name = "LLM Proxy"
base_url = "http://localhost:8891/v1"
requires_openai_auth = true
wire_api = "responses"
`

	cfg := testParse(t, content)
	mp, ok := cfg.ModelProviders["llm-proxy"]
	if !ok {
		t.Fatal("model_providers.llm-proxy not found")
	}
	if mp.Name != "LLM Proxy" {
		t.Errorf("Name = %q", mp.Name)
	}
	if mp.BaseURL != "http://localhost:8891/v1" {
		t.Errorf("BaseURL = %q", mp.BaseURL)
	}
	if mp.RequiresOpenAIAuth == nil || !*mp.RequiresOpenAIAuth {
		t.Error("RequiresOpenAIAuth should be true")
	}
	if mp.WireAPI != "responses" {
		t.Errorf("WireAPI = %q", mp.WireAPI)
	}
}

func TestReadFile_MCPServers(t *testing.T) {
	content := `
[mcp_servers.openaiDeveloperDocs]
url = "https://developers.openai.com/mcp"

[mcp_servers.skynet-base]
command = "skynet-mcp"
args = [ "--mcp=skynet-base" ]
`

	cfg := testParse(t, content)

	s1, ok := cfg.MCPServers["openaiDeveloperDocs"]
	if !ok {
		t.Fatal("mcp_servers.openaiDeveloperDocs not found")
	}
	if s1.URL != "https://developers.openai.com/mcp" {
		t.Errorf("URL = %q", s1.URL)
	}

	s2, ok := cfg.MCPServers["skynet-base"]
	if !ok {
		t.Fatal("mcp_servers.skynet-base not found")
	}
	if s2.Command != "skynet-mcp" {
		t.Errorf("Command = %q", s2.Command)
	}
	if len(s2.Args) != 1 || s2.Args[0] != "--mcp=skynet-base" {
		t.Errorf("Args = %v", s2.Args)
	}
}

func TestReadFile_Projects(t *testing.T) {
	content := `
[projects."/Users/test/project"]
trust_level = "trusted"

[projects."/tmp/sandbox"]
trust_level = "untrusted"
`

	cfg := testParse(t, content)
	if len(cfg.Projects) != 2 {
		t.Fatalf("len(Projects) = %d, want 2", len(cfg.Projects))
	}
	if cfg.Projects["/Users/test/project"].TrustLevel != TrustTrusted {
		t.Errorf("trust_level = %q", cfg.Projects["/Users/test/project"].TrustLevel)
	}
	if cfg.Projects["/tmp/sandbox"].TrustLevel != TrustUntrusted {
		t.Errorf("trust_level = %q", cfg.Projects["/tmp/sandbox"].TrustLevel)
	}
}

func TestReadFile_Plugins(t *testing.T) {
	content := `
[plugins."computer-use@openai-bundled"]
enabled = true

[plugins."browser-use@openai-bundled"]
enabled = false
`

	cfg := testParse(t, content)
	if len(cfg.Plugins) != 2 {
		t.Fatalf("len(Plugins) = %d, want 2", len(cfg.Plugins))
	}

	p1 := cfg.Plugins["computer-use@openai-bundled"]
	if p1.Enabled == nil || !*p1.Enabled {
		t.Error("computer-use should be enabled")
	}

	p2 := cfg.Plugins["browser-use@openai-bundled"]
	if p2.Enabled == nil || *p2.Enabled {
		t.Error("browser-use should be disabled")
	}
}

func TestReadFile_Features(t *testing.T) {
	content := `
[features]
hooks = true
fast_mode = false
memories = true
`

	cfg := testParse(t, content)
	if cfg.Features.Hooks == nil || !*cfg.Features.Hooks {
		t.Error("hooks should be true")
	}
	if cfg.Features.FastMode == nil || *cfg.Features.FastMode {
		t.Error("fast_mode should be false")
	}
	if cfg.Features.Memories == nil || !*cfg.Features.Memories {
		t.Error("memories should be true")
	}
	if cfg.Features.ShellSnapshot != nil {
		t.Error("unset features should be nil")
	}
}

func TestReadFile_Marketplaces(t *testing.T) {
	content := `
[marketplaces.openai-bundled]
last_updated = "2026-04-27T07:49:51Z"
source_type = "local"
source = "/Users/xhd2015/.codex/.tmp/bundled-marketplaces/openai-bundled"
`

	cfg := testParse(t, content)
	m, ok := cfg.Marketplaces["openai-bundled"]
	if !ok {
		t.Fatal("marketplaces.openai-bundled not found")
	}
	if m.SourceType != "local" {
		t.Errorf("SourceType = %q", m.SourceType)
	}
}

func TestReadFile_Agents(t *testing.T) {
	content := `
[agents.reviewer]
description = "Reviews code changes"
nickname_candidates = ["review-bot", "code-checker"]
`

	cfg := testParse(t, content)
	a, ok := cfg.Agents["reviewer"]
	if !ok {
		t.Fatal("agents.reviewer not found")
	}
	if a.Description != "Reviews code changes" {
		t.Errorf("Description = %q", a.Description)
	}
	if len(a.NicknameCandidates) != 2 {
		t.Errorf("len(NicknameCandidates) = %d, want 2", len(a.NicknameCandidates))
	}
}

func TestReadFile_Empty(t *testing.T) {
	cfg := testParse(t, "")
	if cfg.Model != "" {
		t.Errorf("Model = %q, want empty", cfg.Model)
	}
}

func TestReadFile_NotFound(t *testing.T) {
	cfg, err := ReadFile("/nonexistent/path/to/codex-config.toml")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if cfg.Path != "/nonexistent/path/to/codex-config.toml" {
		t.Errorf("Path = %q", cfg.Path)
	}
}

func testParse(t *testing.T, content string) *Config {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	cfg, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return cfg
}
