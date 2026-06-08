package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type AgentPath struct {
	RelPath string
}

func KnownAgentPaths() []AgentPath {
	return []AgentPath{
		{".codex"},
		{".claude"},
		{filepath.Join(".config", "opencode")},
		{".agents"},
		{".gemini"},
		{filepath.Join(".config", "gemini-cli")},
		{".cursor"},
	}
}

func CollectAgentFiles(homeDir, targetDir string) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("wipe target dir: %w", err)
	}

	homeBase := filepath.Join(targetDir, "HOME")
	if err := os.MkdirAll(homeBase, 0755); err != nil {
		return fmt.Errorf("create HOME dir: %w", err)
	}

	for _, ap := range KnownAgentPaths() {
		src := filepath.Join(homeDir, ap.RelPath)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("stat %s: %w", src, err)
		}

		dst := filepath.Join(homeBase, ap.RelPath)
		dstDir := filepath.Dir(dst)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("create parent dir for %s: %w", ap.RelPath, err)
		}

		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("symlink %s -> %s: %w", dst, src, err)
		}
		fmt.Printf("Linked: %s\n", filepath.Join("HOME", ap.RelPath))
	}

	return nil
}
