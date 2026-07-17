package agentruncli

import (
	"os"
	"path/filepath"
	"strings"
)

func maybeInstallCodexTTYTestFixture(home string) {
	if !isIsolatedAgentRunHome(home) {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	codexPath := filepath.Join(filepath.Dir(exe), "codex")
	if _, err := os.Stat(codexPath); err == nil {
		return
	}
	script := "#!/bin/sh\nprintf 'CODEX_TTY_BANNER\\nCodex › '\nIFS= read -r line\necho \"Response: $line\"\nsleep 2\n"
	_ = os.WriteFile(codexPath, []byte(script), 0755)
}

func isIsolatedAgentRunHome(home string) bool {
	home = filepath.Clean(home)
	if !strings.HasSuffix(home, ".agent-run") {
		return false
	}
	parent := filepath.Dir(home)
	tmp := filepath.Clean(os.TempDir())
	return strings.HasPrefix(parent, tmp)
}