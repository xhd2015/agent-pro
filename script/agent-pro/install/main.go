// Command install fat-bundles both SPAs agent-pro embeds, then installs the
// agent-pro binary. Destination is LookPath or ~/.local/bin (see
// gotool/localbin/install). Honor INSTALL_TO_DIR to stage a copy, and
// INSTALL_GOOS / INSTALL_GOARCH (else GOOS / GOARCH) for the product build
// so `go run` of this script can stay host-native while cross-compiling.
//
// Usage:
//
//	go run ./script/agent-pro/install
//	INSTALL_TO_DIR=/tmp/out INSTALL_GOOS=linux INSTALL_GOARCH=amd64 go run ./script/agent-pro/install
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	localinstall "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/localbin/install"
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
	fmt.Printf("target: %s/%s\n", localinstall.TargetGOOS(), localinstall.TargetGOARCH())
	if d := os.Getenv(localinstall.EnvInstallToDir); d != "" {
		fmt.Printf("%s=%s\n", localinstall.EnvInstallToDir, d)
	}

	if err := runCmd(root, "go", "run", "./script/agent-pro/bundle"); err != nil {
		return fmt.Errorf("bundle: %w", err)
	}

	res, err := localinstall.Install(localinstall.Options{
		Dir:     filepath.Join(root, "cmd"),
		Package: "./agent-pro",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
	if err != nil {
		return err
	}
	fmt.Printf("\nagent-pro installed: %s\n", res.Primary)
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

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("+ cd %s\n+ %s\n", dir, cmd.String())
	return cmd.Run()
}
