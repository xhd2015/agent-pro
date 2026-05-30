package podman

import (
	"os/exec"
	"testing"
)

func podmanAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skip("podman not installed, skipping integration test")
	}
	// Check server is reachable
	c := exec.Command("podman", "version", "--format", "{{.Server.Version}}")
	if err := c.Run(); err != nil {
		t.Skip("podman server not reachable, skipping integration test")
	}
}

func TestPodmanArch(t *testing.T) {
	podmanAvailable(t)

	arch, err := PodmanArch()
	if err != nil {
		t.Fatalf("PodmanArch() error: %v", err)
	}
	if arch == "" {
		t.Error("PodmanArch() returned empty string")
	}

	validArches := map[string]bool{"amd64": true, "arm64": true, "arm": true, "386": true}
	if !validArches[arch] {
		t.Logf("PodmanArch() = %s (unexpected but may be valid)", arch)
	}
}

func TestEnsurePodman_Idempotent(t *testing.T) {
	podmanAvailable(t)

	// Calling EnsurePodman twice should succeed both times
	if err := EnsurePodman(); err != nil {
		t.Fatalf("EnsurePodman() first call: %v", err)
	}
	if err := EnsurePodman(); err != nil {
		t.Fatalf("EnsurePodman() second call: %v", err)
	}
}
