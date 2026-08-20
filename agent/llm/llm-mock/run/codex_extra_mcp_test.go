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
	if err := writeCodexConfigToml(dest, 1234, "", nil); err != nil {
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

func TestWriteCodexConfigTomlGeneratedMockMCPThenExtra(t *testing.T) {
	dir := t.TempDir()
	extra := filepath.Join(dir, "extra-mcp.toml")
	if err := os.WriteFile(extra, []byte("[mcp_servers.extra_hang]\ncommand = \"true\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvExtraMCPTOMLFile, extra)
	spec, err := parseMCPSpec("slow_01=1s-10s")
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "config.toml")
	if err := writeCodexConfigToml(dest, 9, "/tmp/mock-mcp", []MCPSpec{spec}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "[mcp_servers.slow_01]") {
		t.Fatalf("missing generated mock-mcp:\n%s", s)
	}
	if !strings.Contains(s, "[mcp_servers.extra_hang]") {
		t.Fatalf("missing extra MCP block:\n%s", s)
	}
	if idxGen, idxExtra := strings.Index(s, "[mcp_servers.slow_01]"), strings.Index(s, "[mcp_servers.extra_hang]"); idxGen < 0 || idxExtra < idxGen {
		t.Fatalf("generated block should precede extra file:\n%s", s)
	}
}
