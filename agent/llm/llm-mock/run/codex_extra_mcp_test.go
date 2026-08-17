package run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteCodexConfigTomlExtraMCPFile(t *testing.T) {
	dir := t.TempDir()
	extra := filepath.Join(dir, "extra-mcp.toml")
	if err := os.WriteFile(extra, []byte("[mcp_servers.slow_01]\ncommand = \"true\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvExtraMCPTOMLFile, extra)
	dest := filepath.Join(dir, "config.toml")
	if err := writeCodexConfigToml(dest, 1234); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "[mcp_servers.slow_01]") {
		t.Fatalf("config.toml missing extra MCP block:\n%s", s)
	}
	if !strings.Contains(s, `base_url = "http://127.0.0.1:1234/v1"`) {
		t.Fatalf("config.toml missing mock provider:\n%s", s)
	}
}
