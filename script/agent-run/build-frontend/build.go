package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
)

const agentRunPlaceholder = `agent-pro frontend-agent-run embed placeholder

This tree is incomplete in git / module zip so bare ` + "`go install`" + ` compiles.
Stage a full SPA via:

  go run ./script/agent-run/build-frontend
  # or
  go run ./script/agent-run/install

At runtime, thin embeds hydrate from GitHub Releases when
AGENT_PRO_ASSET_BASE_URL is set (see doc/assets-hydrate.md).
`

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	if _, err := exec.LookPath("bun"); err != nil {
		return fmt.Errorf("bun is not installed, install it from https://bun.sh/docs/installation")
	}

	// Always install so package.json/lockfile changes are applied even when
	// node_modules already exists but is missing newer deps.
	if err := cmd.Debug().Dir("frontend-agent-run").Run("bun", "install"); err != nil {
		return fmt.Errorf("bun install (frontend-agent-run): %w", err)
	}

	if err := cmd.Debug().Dir("frontend-agent-run").Run("bun", "run", "build"); err != nil {
		return fmt.Errorf("bun run build (frontend-agent-run): %w", err)
	}
	// Vite empties dist/; restore tracked placeholder so git/module stay embed-safe.
	return writePlaceholder("frontend-agent-run/dist/placeholder.txt", agentRunPlaceholder)
}

func writePlaceholder(rel, content string) error {
	if err := os.MkdirAll(filepath.Dir(rel), 0o755); err != nil {
		return err
	}
	return os.WriteFile(rel, []byte(content), 0o644)
}
