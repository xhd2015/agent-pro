package podman

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const DefaultNetworkCheckTimeout = 15 * time.Second

// EnsurePodman checks that podman is installed and the machine is running.
// It will init and start the podman machine if needed.
func EnsurePodman() error {
	fmt.Println("=== Checking podman ===")

	if _, err := exec.LookPath("podman"); err != nil {
		return fmt.Errorf("podman is not installed. Please install it first:\n  macOS: brew install podman\n  Linux: https://podman.io/docs/installation")
	}

	var infoBuf bytes.Buffer
	infoCmd := exec.Command("podman", "machine", "info")
	infoCmd.Stdout = &infoBuf
	infoCmd.Stderr = &infoBuf
	if err := infoCmd.Run(); err != nil {
		fmt.Println("No podman machine found. Initializing...")
		if err := Run("podman", "machine", "init"); err != nil {
			return fmt.Errorf("podman machine init failed: %v", err)
		}
		fmt.Println("Starting podman machine...")
		if err := Run("podman", "machine", "start"); err != nil {
			return fmt.Errorf("podman machine start failed: %v", err)
		}
		fmt.Println("Podman machine started.")
		return nil
	}

	var versionBuf bytes.Buffer
	versionCmd := exec.Command("podman", "version", "--format", "{{.Server.Version}}")
	versionCmd.Stdout = &versionBuf
	versionCmd.Stderr = &versionBuf
	if err := versionCmd.Run(); err != nil {
		fmt.Println("Podman machine is not running. Starting...")
		if startErr := Run("podman", "machine", "start"); startErr != nil {
			return fmt.Errorf("podman machine start failed: %v (original error: %v)", startErr, err)
		}
		fmt.Println("Podman machine started.")
		return nil
	}

	fmt.Printf("Podman ready (server version: %s)\n", strings.TrimSpace(versionBuf.String()))

	if err := CheckVMNetwork(); err != nil {
		return err
	}

	return nil
}

// CheckVMNetwork verifies the podman VM can reach the internet.
// If not, it restarts the machine (a common fix for stale network bridges).
func CheckVMNetwork() error {
	fmt.Print("Checking VM network connectivity... ")

	ok, err := probeVMNetwork()
	if err == nil && ok {
		fmt.Println("OK")
		return nil
	}

	fmt.Println("FAILED (no connectivity)")
	fmt.Println("Restarting podman machine to fix networking...")
	if err := Run("podman", "machine", "stop"); err != nil {
		return fmt.Errorf("podman machine stop failed: %v", err)
	}
	if err := Run("podman", "machine", "start"); err != nil {
		return fmt.Errorf("podman machine start failed: %v", err)
	}

	fmt.Print("Re-checking VM network connectivity... ")
	ok2, err2 := probeVMNetwork()
	if err2 != nil || !ok2 {
		return fmt.Errorf("podman VM still has no network connectivity after restart. Please check your host network and try again")
	}

	fmt.Println("OK")
	return nil
}

func probeVMNetwork() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultNetworkCheckTimeout)
	defer cancel()

	checkCmd := exec.CommandContext(ctx, "podman", "machine", "ssh", "--",
		"ping", "-c", "1", "-W", "5", "8.8.8.8")
	var out bytes.Buffer
	checkCmd.Stdout = &out
	checkCmd.Stderr = &out
	err := checkCmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return false, fmt.Errorf("network check timed out after %v", DefaultNetworkCheckTimeout)
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

// PodmanArch returns the architecture of the podman VM (e.g. "amd64", "arm64").
func PodmanArch() (string, error) {
	var buf bytes.Buffer
	c := exec.Command("podman", "info", "--format", "{{.Host.Arch}}")
	c.Stdout = &buf
	c.Stderr = &buf
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("failed to detect podman architecture: %v", err)
	}
	arch := strings.TrimSpace(buf.String())
	if arch == "" {
		return "", fmt.Errorf("podman returned empty architecture")
	}
	return arch, nil
}

// Run executes a command with stdout/stderr connected to the terminal.
// Use this for long-running or verbose podman commands where output should
// be visible to the user.
func Run(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
