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

func ValidateTargetDir(targetDir string) error {
	fi, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat target dir: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("target exists but is not a directory: %s", targetDir)
	}
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("read target dir: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("target directory is not empty: %s", targetDir)
	}
	return nil
}

func CollectAgentFiles(homeDir, targetDir string) error {
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
