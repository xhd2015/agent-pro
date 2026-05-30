package podman

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// InspectStatus returns the running status of the named container
// (e.g. "running", "stopped"). Returns an error if the container does not exist.
func InspectStatus(containerName string) (string, error) {
	var buf strings.Builder
	c := exec.Command("podman", "inspect", "--format", "{{.State.Status}}", containerName)
	c.Stdout = &buf
	c.Stderr = &buf
	if err := c.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

// InspectLabel reads a label value from the container's config.
// Returns "" if the label is not set or the container does not exist.
func InspectLabel(containerName, label string) string {
	var buf strings.Builder
	c := exec.Command("podman", "inspect", "--format",
		fmt.Sprintf("{{index .Config.Labels %q}}", label), containerName)
	c.Stdout = &buf
	if err := c.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

// Remove stops and removes a container. If force is true, uses -f.
// Returns silently if the container does not exist.
func Remove(containerName string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerName)
	return Run("podman", args...)
}

// Stop stops a running container.
func Stop(containerName string) error {
	return Run("podman", "stop", containerName)
}

// Start starts a stopped container.
func Start(containerName string) error {
	return Run("podman", "start", containerName)
}

// CopyTo copies a file or directory from the host into the container.
func CopyTo(containerName, hostPath, containerPath string) error {
	return Run("podman", "cp", hostPath, containerName+":"+containerPath)
}

// Exec executes a command inside a running container with terminal I/O.
func Exec(containerName string, cmdAndArgs ...string) error {
	args := append([]string{"exec", containerName}, cmdAndArgs...)
	c := exec.Command("podman", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// ExecOutput executes a command inside a running container and returns stdout.
func ExecOutput(containerName string, cmdAndArgs ...string) (string, error) {
	args := append([]string{"exec", containerName}, cmdAndArgs...)
	var buf strings.Builder
	c := exec.Command("podman", args...)
	c.Stdout = &buf
	c.Stderr = &buf
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("podman exec failed: %v (stderr: %s)", err, buf.String())
	}
	return strings.TrimSpace(buf.String()), nil
}

// Create creates a new container with the given arguments.
// args should include all podman create arguments (e.g. "--name", "my-container", "image", ...).
func Create(args ...string) error {
	allArgs := append([]string{"create"}, args...)
	return Run("podman", allArgs...)
}

// OnSignal runs fn when SIGINT or SIGTERM is received.
// Returns a cancel function to stop listening.
func OnSignal(fn func()) (cancel func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		select {
		case <-sigChan:
			fn()
		case <-done:
		}
	}()
	return func() {
		signal.Stop(sigChan)
		close(done)
	}
}

// FollowLogs streams container logs to stdout/stderr until the context is
// cancelled or the container exits.
func FollowLogs(ctx context.Context, containerName string) error {
	logsCmd := exec.CommandContext(ctx, "podman", "logs", "-f", containerName)
	logsCmd.Stdout = os.Stdout
	logsCmd.Stderr = os.Stderr
	if err := logsCmd.Start(); err != nil {
		return fmt.Errorf("failed to follow container logs: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- logsCmd.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = exec.Command("podman", "stop", containerName).Run()
	case err := <-done:
		if err != nil && ctx.Err() == nil {
			return fmt.Errorf("container exited with error: %v", err)
		}
	}

	return nil
}
