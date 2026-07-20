// Command bundle builds fat SPA embeds required by agent-pro:
//   - frontend/           (traces viewer)
//   - frontend-agent-run/ (grok session view --web message cards)
//
// Usage:
//
//	go run ./script/agent-pro/bundle
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := moduleRoot()
	if err != nil {
		return fmt.Errorf("resolve module root: %w", err)
	}
	fmt.Printf("module root: %s\n", root)

	if err := buildAgentProFrontend(root); err != nil {
		return err
	}
	if err := buildAgentRunFrontend(root); err != nil {
		return err
	}

	fmt.Println("\nbundle: frontend/dist staged (agent-pro SPA)")
	fmt.Println("bundle: frontend-agent-run/dist staged (agent-run SPA for grok view --web)")
	return nil
}

func moduleRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for i := 0; i < 10; i++ {
		if looksLikeRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not locate module root from %s", filepath.Dir(thisFile))
}

func looksLikeRoot(dir string) bool {
	for _, rel := range []string{
		"go.mod",
		"frontend",
		"frontend-agent-run",
		filepath.Join("cmd", "agent-pro"),
		filepath.Join("script", "build-frontend"),
		filepath.Join("script", "agent-run", "build-frontend"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			return false
		}
	}
	return true
}

func buildAgentProFrontend(root string) error {
	fmt.Println("\n== build agent-pro frontend ==")
	return runCmd(root, "go", "run", "./script/build-frontend")
}

func buildAgentRunFrontend(root string) error {
	fmt.Println("\n== build agent-run frontend ==")
	return runCmd(root, "go", "run", "./script/agent-run/build-frontend")
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("+ cd %s\n+ %s\n", dir, cmd.String())
	return cmd.Run()
}
