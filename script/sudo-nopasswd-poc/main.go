// Conceptual POC for passwordless-sudo setup used by ws-proxy vpn.
//
// Models the future --no-setup-sudo / auto-setup flow with a stand-in "hello"
// command instead of sing-box. Validates detect → setup → reuse → remove.
//
// Usage:
//
//	go run ./script/sudo-nopasswd-poc detect
//	go run ./script/sudo-nopasswd-poc auto-setup
//	go run ./script/sudo-nopasswd-poc remove-sudo
//
// Manual validation:
//  1. remove-sudo          # start clean (also flushes sudo timestamp cache)
//  2. auto-setup           # first run: sudoers install prompts once, prints hello
//  3. auto-setup           # second run (or after 10+ min): no password, prints hello
//  4. remove-sudo          # tear down
//  5. auto-setup           # prompts for password again
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/agent-pro/pkgs/sudosetup"
)

const (
	cacheDirName = "remote-agent-sudo-poc"
	helloScript  = "hello.sh"
	sudoersName  = "remote-agent-sudo-poc"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "detect":
		os.Exit(runDetect())
	case "auto-setup":
		if err := runAutoSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "auto-setup: %v\n", err)
			os.Exit(1)
		}
	case "remove-sudo":
		if err := runRemoveSudo(); err != nil {
			fmt.Fprintf(os.Stderr, "remove-sudo: %v\n", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	mgr := manager()
	fmt.Printf(`sudo-nopasswd-poc — conceptual NOPASSWD sudoers flow for ws-proxy vpn

Usage:
  go run ./script/sudo-nopasswd-poc <command>

Commands:
  detect       Report installed NOPASSWD rule vs sudo timestamp cache
  auto-setup   Ensure NOPASSWD rule exists, then run hello (setup prompts once)
  remove-sudo  Delete sudoers drop-in, local manifest, and flush sudo cache

Files:
  hello script    ~/.cache/%s/%s
  install manifest %s
  sudoers drop-in %s

Expected manual test:
  remove-sudo → auto-setup (password once) → auto-setup (no password) → remove-sudo → auto-setup (password again)
`, cacheDirName, helloScript, mgr.ManifestPath(), mgr.SudoersPath())
}

func runDetect() int {
	mgr, err := managerWithHello()
	if err != nil {
		fmt.Fprintf(os.Stderr, "detect: %v\n", err)
		return 1
	}

	status := mgr.Detect()
	fmt.Printf("detect: NOPASSWD rule installed: %s\n", boolLabel(status.Installed))
	if status.InstallDetail != "" {
		fmt.Printf("detect:   %s\n", status.InstallDetail)
	}
	fmt.Printf("detect: sudo timestamp cache warm: %s\n", boolLabel(status.CacheWarm))
	fmt.Printf("detect: hello via sudo -n right now: %s\n", boolLabel(status.CanRunNonInteractive))
	if status.CanRunDetail != "" {
		fmt.Printf("detect:   %s\n", status.CanRunDetail)
	}
	fmt.Printf("detect: verdict: %s\n", status.Verdict)

	if status.Installed && status.CanRunNonInteractive {
		return 0
	}
	return 1
}

func runAutoSetup() error {
	helloPath, err := ensureHelloScript()
	if err != nil {
		return err
	}

	mgr := managerForHello(helloPath)
	if installed, _ := mgr.IsInstalled(); installed {
		fmt.Println("auto-setup: NOPASSWD rule already installed, skipping setup")
		return runHello(helloPath, false)
	}

	fmt.Println("auto-setup: NOPASSWD rule not installed; installing sudoers drop-in")
	fmt.Printf("auto-setup: will allow: sudo %s\n", helloPath)
	fmt.Println("auto-setup: you may be prompted for your login password once to install the rule")

	if err := mgr.EnsureInstalled(); err != nil {
		return err
	}

	if installed, detail := mgr.IsInstalled(); !installed {
		return fmt.Errorf("setup finished but rule not recorded as installed: %s", detail)
	}
	fmt.Println("auto-setup: sudoers drop-in installed and recorded")

	status := mgr.Detect()
	if !status.CanRunNonInteractive {
		return fmt.Errorf("setup finished but sudo -n hello failed: %s", status.CanRunDetail)
	}

	return runHello(helloPath, false)
}

func runRemoveSudo() error {
	mgr, err := managerWithHello()
	if err != nil {
		return err
	}

	sudoersPath := mgr.SudoersPath()
	removedDropIn := false
	if _, err := os.Stat(sudoersPath); err == nil {
		fmt.Printf("remove-sudo: deleting %s (sudo may prompt once)\n", sudoersPath)
		removedDropIn = true
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", sudoersPath, err)
	}

	manifestPath := mgr.ManifestPath()
	removedManifest := false
	if _, err := os.Stat(manifestPath); err == nil {
		removedManifest = true
	}

	if err := mgr.Remove(); err != nil {
		return err
	}

	helloPath, err := helloScriptPath()
	if err == nil {
		_ = os.Remove(helloPath)
	}

	switch {
	case removedDropIn:
		fmt.Println("remove-sudo: done — sudoers drop-in removed and timestamp cache flushed")
	case removedManifest:
		fmt.Println("remove-sudo: done — local manifest removed and timestamp cache flushed")
	default:
		fmt.Println("remove-sudo: nothing to remove; timestamp cache flushed")
	}
	fmt.Println("remove-sudo: next auto-setup should prompt for password again")
	return nil
}

func manager() *sudosetup.Manager {
	return &sudosetup.Manager{
		Config: sudosetup.Config{
			CacheDirName: cacheDirName,
			SudoersName:  sudoersName,
		},
	}
}

func managerForHello(helloPath string) *sudosetup.Manager {
	mgr := manager()
	mgr.Rule = sudosetup.Rule{Command: helloPath}
	return mgr
}

func managerWithHello() (*sudosetup.Manager, error) {
	helloPath, err := helloScriptPath()
	if err != nil {
		return nil, err
	}
	return managerForHello(helloPath), nil
}

func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, cacheDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func helloScriptPath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, helloScript), nil
}

func ensureHelloScript() (string, error) {
	path, err := helloScriptPath()
	if err != nil {
		return "", err
	}
	content := "#!/bin/sh\necho hello\n"
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		return "", fmt.Errorf("write hello script: %w", err)
	}
	return path, nil
}

func runHello(helloPath string, allowPrompt bool) error {
	var cmd *exec.Cmd
	if allowPrompt {
		cmd = exec.Command("sudo", helloPath)
		cmd.Stdin = os.Stdin
	} else {
		cmd = exec.Command("sudo", "-n", helloPath)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run hello: %w", err)
	}
	return nil
}

func boolLabel(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}