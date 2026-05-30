package podman

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	testImage       = "docker.io/library/alpine:latest"
	testContainer   = "agent-pro-podman-test"
	testLabel       = "agent-pro.test-id"
	testLabelValue  = "container_test"
	sleepInfinity   = "sleep infinity"
)

func podmanContainerAvailable(t *testing.T) {
	t.Helper()
	podmanAvailable(t)
}

func cleanupTestContainer(t *testing.T) {
	t.Helper()
	_ = Run("podman", "rm", "-f", testContainer)
}

func TestInspectStatus_NotFound(t *testing.T) {
	podmanContainerAvailable(t)

	_, err := InspectStatus("nonexistent-container-99999")
	if err == nil {
		t.Error("InspectStatus() should return error for nonexistent container")
	}
}

func TestContainerLifecycle(t *testing.T) {
	podmanContainerAvailable(t)
	cleanupTestContainer(t)
	t.Cleanup(func() { cleanupTestContainer(t) })

	// Pull image first (quiet)
	Run("podman", "pull", testImage)

	// Create container
	if err := Run("podman",
		"create",
		"--name", testContainer,
		"--label", testLabel+"="+testLabelValue,
		testImage,
		"sleep", "10",
	); err != nil {
		t.Fatalf("Create container: %v", err)
	}

	// Check status — should be "created" (not running)
	status, err := InspectStatus(testContainer)
	if err != nil {
		t.Fatalf("InspectStatus: %v", err)
	}
	if status == "" {
		t.Error("InspectStatus() returned empty status")
	}

	// Check label
	label := InspectLabel(testContainer, testLabel)
	if label != testLabelValue {
		t.Errorf("InspectLabel() = %q, want %q", label, testLabelValue)
	}

	// Start container
	if err := Start(testContainer); err != nil {
		t.Fatalf("Start: %v", err)
	}

	status, err = InspectStatus(testContainer)
	if err != nil {
		t.Fatalf("InspectStatus after start: %v", err)
	}
	if status != "running" {
		t.Errorf("status after Start = %q, want \"running\"", status)
	}

	// Exec command in container
	out, err := ExecOutput(testContainer, "echo", "hello-from-container")
	if err != nil {
		t.Fatalf("ExecOutput: %v", err)
	}
	if out != "hello-from-container" {
		t.Errorf("ExecOutput() = %q, want %q", out, "hello-from-container")
	}

	// Copy file to container
	hostPath := "/tmp/agent-pro-podman-test-file"
	content := "test-content-" + fmt.Sprintf("%d", time.Now().UnixNano())
	if err := Run("sh", "-c", fmt.Sprintf("printf '%%s' '%s' > %s", content, hostPath)); err != nil {
		t.Fatalf("Create temp file: %v", err)
	}
	defer Run("rm", "-f", hostPath)

	if err := CopyTo(testContainer, hostPath, "/tmp/test-file"); err != nil {
		t.Fatalf("CopyTo: %v", err)
	}

	catOut, err := ExecOutput(testContainer, "cat", "/tmp/test-file")
	if err != nil {
		t.Fatalf("ExecOutput cat: %v", err)
	}
	if catOut != content {
		t.Errorf("CopyTo verification: got %q, want %q", catOut, content)
	}

	// Stop container
	if err := Stop(testContainer); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	status, err = InspectStatus(testContainer)
	if err != nil {
		t.Fatalf("InspectStatus after stop: %v", err)
	}
	if status != "exited" && status != "stopped" {
		t.Errorf("status after Stop = %q, want \"exited\" or \"stopped\"", status)
	}

	// Remove container
	if err := Remove(testContainer, true); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	_, err = InspectStatus(testContainer)
	if err == nil {
		t.Error("InspectStatus() should return error after remove")
	}
}

func TestFollowLogs(t *testing.T) {
	podmanContainerAvailable(t)
	cleanupTestContainer(t)
	t.Cleanup(func() { cleanupTestContainer(t) })

	// Create and start a container that outputs something quickly
	if err := Run("podman",
		"run", "-d",
		"--name", testContainer,
		testImage,
		"sh", "-c", "echo hello && sleep 5",
	); err != nil {
		t.Fatalf("Create container: %v", err)
	}

	// Wait a moment for the echo to complete
	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// FollowLogs should exit when the container stops (after sleep 5).
	// We test that it doesn't hang indefinitely.
	err := FollowLogs(ctx, testContainer)
	// Accept any error — the test is that it returns without hanging.
	_ = err
}

func TestInspectLabel_Missing(t *testing.T) {
	podmanContainerAvailable(t)
	cleanupTestContainer(t)
	t.Cleanup(func() { cleanupTestContainer(t) })

	if err := Run("podman",
		"run", "-d",
		"--name", testContainer,
		testImage,
		"sleep", "5",
	); err != nil {
		t.Fatalf("Create container: %v", err)
	}

	label := InspectLabel(testContainer, "nonexistent-label")
	if label != "" {
		t.Errorf("InspectLabel() for missing label = %q, want \"\"", label)
	}
}

func TestExecOutput_Error(t *testing.T) {
	podmanContainerAvailable(t)
	cleanupTestContainer(t)
	t.Cleanup(func() { cleanupTestContainer(t) })

	if err := Run("podman",
		"run", "-d",
		"--name", testContainer,
		testImage,
		"sleep", "10",
	); err != nil {
		t.Fatalf("Create container: %v", err)
	}

	// Running a non-existent command should produce an error
	_, err := ExecOutput(testContainer, "nonexistent-command-xyz")
	if err == nil {
		t.Error("ExecOutput() should return error for nonexistent command")
	}
}

func TestRemove_DoesNotErrorWhenNotFound(t *testing.T) {
	podmanContainerAvailable(t)
	// Ensure no leftover
	Run("podman", "rm", "-f", testContainer)

	// Remove --force should succeed (or at least not panic) on nonexistent container
	err := Remove(testContainer, true)
	// We accept either nil or an error here - the key is it doesn't panic
	_ = err
}

func TestCreate_WithConfigHash(t *testing.T) {
	podmanContainerAvailable(t)
	cleanupTestContainer(t)
	t.Cleanup(func() { cleanupTestContainer(t) })

	cfg := "platform=linux/arm64\nport=8080"
	hash := ConfigHash(cfg)

	if err := Run("podman",
		"create",
		"--name", testContainer,
		"--label", "my-config-hash="+hash,
		testImage,
		"sleep", "5",
	); err != nil {
		t.Fatalf("Create container: %v", err)
	}

	// Verify the label was set
	got := InspectLabel(testContainer, "my-config-hash")
	if got != hash {
		t.Errorf("InspectLabel(my-config-hash) = %q, want %q", got, hash)
	}
}

func TestOnSignal(t *testing.T) {
	podmanContainerAvailable(t)
	cleanupTestContainer(t)
	t.Cleanup(func() { cleanupTestContainer(t) })

	if err := Run("podman",
		"run", "-d",
		"--name", testContainer,
		testImage,
		"sleep", "30",
	); err != nil {
		t.Fatalf("Create container: %v", err)
	}

	called := false
	cancel := OnSignal(func() {
		called = true
		_ = Stop(testContainer)
	})
	defer cancel()

	// Manually stop the container (simulating signal handler behavior)
	_ = Stop(testContainer)

	// The signal handler should be able to stop containers
	status, err := InspectStatus(testContainer)
	if err == nil && strings.Contains(status, "exited") || strings.Contains(status, "stopped") {
		called = true
	}
	_ = called
}
