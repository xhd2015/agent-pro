// Command install is a compatibility entrypoint that delegates to
// script/agent-pro/install (fat both SPAs + go install ./cmd/agent-pro).
//
// Usage:
//
//	go run ./script/install
//	go run ./script/agent-pro/install   # preferred
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

	fmt.Println("Delegating to script/agent-pro/install (both SPAs + agent-pro binary)...")
	cmd := exec.Command("go", "run", "./script/agent-pro/install")
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findRootDir() (string, error) {
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
			return os.Getwd()
		}
		dir = parent
	}
}
