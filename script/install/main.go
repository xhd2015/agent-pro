package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "install: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	rootDir, err := findRootDir()
	if err != nil {
		return err
	}

	fmt.Println("Building frontend...")
	buildFrontend := exec.Command("go", "run", "./script/build-frontend")
	buildFrontend.Dir = rootDir
	buildFrontend.Stdout = os.Stdout
	buildFrontend.Stderr = os.Stderr
	if err := buildFrontend.Run(); err != nil {
		return fmt.Errorf("build frontend: %w", err)
	}

	fmt.Println("Installing agent-pro...")
	installCmd := exec.Command("go", "install", "./cmd/agent-pro")
	installCmd.Dir = rootDir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("install agent-pro: %w", err)
	}

	fmt.Println("agent-pro installed successfully")
	return nil
}

func findRootDir() (string, error) {
	// go run ./script/install resolves the module root automatically.
	// If called with GOPATH mode, walk up to find go.mod.
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, use cwd
			return os.Getwd()
		}
		dir = parent
	}
}
