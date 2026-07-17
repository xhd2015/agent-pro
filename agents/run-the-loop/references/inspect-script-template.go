// Template for script/debug/<name>/main.go
// Copy and adapt — do not import this file directly.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// inspect emits a structured result and exits 0 (pass) or 1 (fail).
func inspect(check string, pass bool, evidence, reason string) {
	result := "FAIL"
	if pass {
		result = "PASS"
	}
	fmt.Printf("CHECK: %s\n", check)
	if evidence != "" {
		fmt.Printf("EVIDENCE: %s\n", evidence)
	}
	fmt.Printf("RESULT: %s\n", result)
	if !pass && reason != "" {
		fmt.Printf("REASON: %s\n", reason)
	}
	if !pass {
		os.Exit(1)
	}
}

func outDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	return filepath.Join(dir, "out")
}

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func main() {
	_ = os.MkdirAll(outDir(), 0o755)

	// --- Example: CLI exit code + output substring ---
	// out, err := run("my-cli", "status")
	// inspect("cli reports ready",
	// 	err == nil && strings.Contains(out, "ready"),
	// 	strings.TrimSpace(out),
	// 	"expected 'ready' in output")

	// --- Example: go test gate ---
	// out, err := run("go", "test", "./pkg/...")
	// inspect("unit tests pass", err == nil, out, "go test failed")

	// --- Example: gh workflow (requires gh auth) ---
	// out, err := run("gh", "run", "list", "--workflow", "CI", "--limit", "1", "--json", "conclusion", "-q", ".[0].conclusion")
	// inspect("latest CI run succeeded",
	// 	err == nil && strings.TrimSpace(out) == "success",
	// 	out, "workflow not success")

	// --- Example: delegate to playwright-debug for screenshot ---
	// screenshot := filepath.Join(outDir(), "screenshot.png")
	// script := fmt.Sprintf(`await page.goto("http://localhost:3000"); await page.screenshot({path: %q}); console.log(await page.title());`, screenshot)
	// out, err := run("playwright-debug", "-e", script)
	// inspect("dashboard title visible",
	// 	err == nil && strings.Contains(out, "Dashboard"),
	// 	screenshot, "title mismatch or screenshot failed")

	// Replace with real checks for your goal:
	inspect("placeholder — replace with real check", false, "", "inspect script not yet implemented")
}