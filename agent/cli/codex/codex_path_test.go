package codex

import (
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/exec"
)

func TestFindExecutableCodex_prefersFirstCandidate(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "codex-a")
	second := filepath.Join(dir, "codex-b")
	for _, p := range []string{first, second} {
		if err := os.WriteFile(p, []byte{0}, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	got, ok := findExecutableCodex([]string{first, second})
	if !ok || got != first {
		t.Fatalf("findExecutableCodex = (%q, %v), want (%q, true)", got, ok, first)
	}
}

func TestProbeCodexInstallPath_findsUnderHome(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "go", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}
	got, ok := findExecutableCodex(codexInstallCandidates(home))
	if !ok || got != codexPath {
		t.Fatalf("findExecutableCodex = (%q, %v), want (%q, true)", got, ok, codexPath)
	}
}

func TestFindAgentPath_probesWhenPATHEmpty(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "go", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	env := exec.NewEnv(&exec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")

	got, err := FindAgentPath(env)
	if err != nil {
		t.Fatalf("FindAgentPath: %v", err)
	}
	if got != codexPath {
		t.Fatalf("FindAgentPath = %q, want %q", got, codexPath)
	}
}

func TestProbeCodexInstallPath_nvmDefault(t *testing.T) {
	home := t.TempDir()
	version := "v20.10.0"
	aliasDir := filepath.Join(home, ".nvm", "alias")
	if err := os.MkdirAll(aliasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aliasDir, "default"), []byte(version), 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(home, ".nvm", "versions", "node", version, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	got, ok := probeCodexInstallPath()
	if !ok || got != codexPath {
		t.Fatalf("probeCodexInstallPath = (%q, %v), want (%q, true)", got, ok, codexPath)
	}
}

func TestProbeCodexInstallPath_npmPrefix(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "npm-global")
	binDir := filepath.Join(prefix, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}

	npmBinDir := filepath.Join(home, ".npm-global", "bin")
	if err := os.MkdirAll(npmBinDir, 0o755); err != nil {
		t.Fatal(err)
	}
	npmScript := filepath.Join(npmBinDir, "npm")
	script := "#!/bin/sh\necho \"" + prefix + "\"\n"
	if err := os.WriteFile(npmScript, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	got, ok := probeCodexInstallPath()
	if !ok || got != codexPath {
		t.Fatalf("probeCodexInstallPath = (%q, %v), want (%q, true)", got, ok, codexPath)
	}
}

func TestRunNpmPrefixGlobal_nodeFallback(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "npm-global")
	binDir := filepath.Join(prefix, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	toolDir := filepath.Join(home, "installed", "node-test", "bin")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	realNode, err := osexec.LookPath("node")
	if err != nil {
		t.Skipf("node not on PATH: %v", err)
	}
	nodeBin := filepath.Join(toolDir, "node")
	if err := os.Symlink(realNode, nodeBin); err != nil {
		t.Fatal(err)
	}

	npmDir := filepath.Join(home, "fake-npm-bin")
	if err := os.MkdirAll(npmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	npmCLI := filepath.Join(npmDir, "npm-cli.js")
	cliScript := "console.log(" + strconvQuote(prefix) + ")\n"
	if err := os.WriteFile(npmCLI, []byte(cliScript), 0o644); err != nil {
		t.Fatal(err)
	}
	npmBin := filepath.Join(npmDir, "npm")
	if err := os.Symlink("npm-cli.js", npmBin); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	got, err := runNpmPrefixGlobal(npmBin, home)
	if err != nil {
		t.Fatalf("runNpmPrefixGlobal: %v", err)
	}
	if got != prefix {
		t.Fatalf("prefix = %q, want %q", got, prefix)
	}
}

func strconvQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func TestResolveNvmVersion_aliasChain(t *testing.T) {
	home := t.TempDir()
	aliasDir := filepath.Join(home, ".nvm", "alias")
	nodeAliasDir := filepath.Join(aliasDir, "node")
	if err := os.MkdirAll(nodeAliasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aliasDir, "default"), []byte("20\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeAliasDir, "20"), []byte("v24.10.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, ok := resolveNvmVersion(home)
	if !ok || got != "v24.10.0" {
		t.Fatalf("resolveNvmVersion = (%q, %v), want (v24.10.0, true)", got, ok)
	}
}

func TestProbeCodexInstallPath_nvmAliasChain(t *testing.T) {
	home := t.TempDir()
	version := "v24.10.0"
	aliasDir := filepath.Join(home, ".nvm", "alias")
	nodeAliasDir := filepath.Join(aliasDir, "node")
	if err := os.MkdirAll(nodeAliasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aliasDir, "default"), []byte("20\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeAliasDir, "20"), []byte(version+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(home, ".nvm", "versions", "node", version, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	got, ok := probeCodexInstallPath()
	if !ok || got != codexPath {
		t.Fatalf("probeCodexInstallPath = (%q, %v), want (%q, true)", got, ok, codexPath)
	}
}

func TestParseNpmConfigPrefix(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "custom-npm-prefix")
	if err := os.WriteFile(filepath.Join(home, ".npmrc"), []byte("prefix="+prefix+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := parseNpmConfigPrefix(home)
	if !ok || got != prefix {
		t.Fatalf("parseNpmConfigPrefix = (%q, %v), want (%q, true)", got, ok, prefix)
	}
}

func TestProbeCodexInstallPath_npmrcPrefix(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "custom-npm-prefix")
	binDir := filepath.Join(prefix, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".npmrc"), []byte("prefix="+prefix+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	got, ok := probeCodexInstallPath()
	if !ok || got != codexPath {
		t.Fatalf("probeCodexInstallPath = (%q, %v), want (%q, true)", got, ok, codexPath)
	}
}

func TestProbeCodexViaLoginShell(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, "custom", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codexPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}
	profile := "export PATH=\"" + binDir + ":$PATH\"\n"
	if err := os.WriteFile(filepath.Join(home, ".bash_profile"), []byte(profile), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("USER", "testuser")
	t.Setenv("PATH", "/usr/bin:/bin")

	got, ok := probeCodexViaLoginShell(home)
	if !ok || got != codexPath {
		t.Fatalf("probeCodexViaLoginShell = (%q, %v), want (%q, true)", got, ok, codexPath)
	}
}