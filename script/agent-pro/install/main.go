// Command install fat-bundles both SPAs agent-pro embeds, then installs the
// agent-pro binary into $GOPATH/bin (or $GOBIN).
//
// Usage:
//
//	go run ./script/agent-pro/install
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

	if err := runCmd(root, nil, "go", "run", "./script/agent-pro/bundle"); err != nil {
		return fmt.Errorf("bundle: %w", err)
	}

	binDir, err := goPathBinDir(root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create GOPATH bin dir %s: %w", binDir, err)
	}
	if err := goInstallAgentPro(root, binDir); err != nil {
		return err
	}

	fmt.Printf("\nagent-pro installed: %s\n", filepath.Join(binDir, "agent-pro"))
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
		filepath.Join("script", "agent-pro", "bundle"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			return false
		}
	}
	return true
}

func goPathBinDir(root string) (string, error) {
	cmd := exec.Command("go", "env", "GOPATH")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return "", fmt.Errorf("go env GOPATH: %w: %s", err, detail)
		}
		return "", fmt.Errorf("go env GOPATH: %w", err)
	}
	gopath := strings.TrimSpace(string(out))
	if gopath == "" {
		return "", fmt.Errorf("go env GOPATH returned empty output")
	}
	return filepath.Join(gopath, "bin"), nil
}

func goInstallAgentPro(root, binDir string) error {
	fmt.Println("\n== go install -C cmd ./agent-pro ==")
	env := upsertEnv(os.Environ(), "GOBIN", binDir)
	if err := runCmd(root, env, "go", "install", "-C", "cmd", "./agent-pro"); err != nil {
		return fmt.Errorf("go install -C cmd ./agent-pro: %w", err)
	}
	return nil
}

func runCmd(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if env != nil {
		cmd.Env = env
	}
	fmt.Printf("+ cd %s\n+ %s\n", dir, cmd.String())
	return cmd.Run()
}

func upsertEnv(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	replaced := false
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			if !replaced {
				out = append(out, prefix+value)
				replaced = true
			}
			continue
		}
		out = append(out, entry)
	}
	if !replaced {
		out = append(out, prefix+value)
	}
	return out
}
