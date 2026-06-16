## Preconditions
- Podman must be installed and the podman machine running.
- Real agent config files must exist on the host for the `Agent` being tested.
- The agent binary must be obtainable (download from GitHub releases or npm).
- An internet connection is required for pulling the container image and downloading binaries.

## Steps
1. Check podman availability; skip if not found.
2. Check that real config files exist for the agent on the host; skip if none.
3. Export configs via `agentconfig.Export` to a temporary zip file.
4. Create and start a Debian container with `sleep infinity` to keep it alive.
5. Install prerequisites (`curl`, `unzip`, and optionally `nodejs` + `npm` for pi).
6. Install the agent binary in the container.
7. Copy the exported zip into the container.
8. Extract the zip entries to the correct agent config directories.
9. Run the agent query with a timeout, capture stdout.
10. Clean up: stop and remove the container (deferred).

## Context
- The real home directory (`os.UserHomeDir()`) is used as the config source.
- Container name is: `agent-pro-test-{agent}-{unixTimestamp}`.
- Each leaf sets a different `Agent` and `Query` in its Setup.
- The `configsExist` helper checks for real config files without modifying them.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	agentconfig "github.com/xhd2015/agent-pro/pkgs/agentconfig"
	"github.com/xhd2015/agent-pro/pkgs/containers/podman"
)

type Request struct {
	Agent          string
	Query          string
	ContainerImage string
	TimeoutSeconds int
}

type Response struct {
	Output string
	Err    error
}

func Setup(t *testing.T, req *Request) error {
	podmanTests := os.Getenv("AGENT_PRO_PODMAN_TESTS")
	if podmanTests != "1" && podmanTests != "true" {
		t.Skipf("AGENT_PRO_PODMAN_TESTS=%q; set AGENT_PRO_PODMAN_TESTS=1 to run podman integration tests", podmanTests)
	}
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skipf("WARNING: podman not found in PATH: %v", err)
	}
	if req.ContainerImage == "" {
		req.ContainerImage = "docker.io/library/debian:bookworm-slim"
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 120
	}
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Logf("[1/9] ensuring podman...")
	if err := podman.EnsurePodman(); err != nil {
		t.Skipf("WARNING: podman not available, skipping podman integration test: %v", err)
	}
	t.Logf("[2/9] homeDir=%s, checking %s configs...", os.Getenv("HOME"), req.Agent)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	if !configsExist(req.Agent, homeDir) {
		t.Skipf("no %s config files found in %s", req.Agent, homeDir)
	}
	t.Logf("[3/9] exporting %s configs from %s...", req.Agent, homeDir)

	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("agent-pro-export-%d.zip", time.Now().UnixNano()))
	defer os.Remove(zipPath)

	if err := agentconfig.Export(req.Agent, homeDir, zipPath); err != nil {
		return nil, fmt.Errorf("export %s configs: %w", req.Agent, err)
	}
	t.Logf("[4/9] zip exported to %s, creating container %s...", zipPath, req.ContainerImage)

	containerName := fmt.Sprintf("agent-pro-test-%s-%d", req.Agent, time.Now().Unix())
	t.Logf("[4/9] container name: %s", containerName)
	defer podman.Remove(containerName, true)

	if err := podman.Create("--name", containerName, req.ContainerImage, "sleep", "infinity"); err != nil {
		t.Skipf("WARNING: create container failed, skipping podman integration test: %v", err)
	}
	t.Logf("[5/9] starting container...")
	if err := podman.Start(containerName); err != nil {
		t.Skipf("WARNING: start container failed, skipping podman integration test: %v", err)
	}

	t.Logf("[6/9] installing prerequisites (apt-get update + curl, unzip)...")
	prereqs := "apt-get update && apt-get install -y curl unzip"
	if req.Agent == "pi" {
		prereqs += " nodejs npm"
	}
	if out, err := podman.ExecOutput(containerName, "sh", "-c", prereqs); err != nil {
		t.Skipf("WARNING: install prerequisites failed, skipping podman integration test: %v (output: %s)", err, out)
	}

	t.Logf("[7/9] installing %s binary...", req.Agent)
	switch req.Agent {
	case "opencode":
		arch, archErr := podman.PodmanArch()
		if archErr != nil {
			t.Skipf("WARNING: detect arch failed, skipping podman integration test: %v", archErr)
		}
		downloadURL := fmt.Sprintf("https://github.com/anthropics/claude-code/releases/latest/download/opencode-linux-%s", arch)
		t.Logf("[7/9] downloading opencode from %s", downloadURL)
		installCmd := fmt.Sprintf("curl -fsSL %s -o /usr/local/bin/opencode && chmod +x /usr/local/bin/opencode", downloadURL)
		if out, installErr := podman.ExecOutput(containerName, "sh", "-c", installCmd); installErr != nil {
			t.Skipf("opencode binary download failed: %v (output: %s)", installErr, out)
		}
	case "pi":
		installCmd := "npm install -g --ignore-scripts @earendil-works/pi-coding-agent && (which pi || (echo 'ERROR: pi binary not found after npm install' && false))"
		t.Logf("[7/9] running: %s", installCmd)
		if out, installErr := podman.ExecOutput(containerName, "sh", "-c", installCmd); installErr != nil {
			t.Skipf("pi npm install failed: %v (output: %s)", installErr, out)
		}
	case "crush":
		arch, archErr := podman.PodmanArch()
		if archErr != nil {
			t.Skipf("WARNING: detect arch failed, skipping podman integration test: %v", archErr)
		}
		downloadURL := fmt.Sprintf("https://github.com/anthropics/claude-code/releases/latest/download/crush-linux-%s", arch)
		t.Logf("[7/9] downloading crush from %s", downloadURL)
		installCmd := fmt.Sprintf("curl -fsSL %s -o /usr/local/bin/crush && chmod +x /usr/local/bin/crush", downloadURL)
		if out, installErr := podman.ExecOutput(containerName, "sh", "-c", installCmd); installErr != nil {
			t.Skipf("crush binary download failed: %v (output: %s)", installErr, out)
		}
	}

	t.Logf("[8/9] copying zip into container...")
	containerZipPath := "/tmp/config.zip"
	if err := podman.CopyTo(containerName, zipPath, containerZipPath); err != nil {
		return nil, fmt.Errorf("copy zip to container: %w", err)
	}

	t.Logf("[9/9] extracting config files in container...")
	extractCmds := []string{
		"mkdir -p /tmp/unpacked",
		"unzip -o /tmp/config.zip -d /tmp/unpacked",
	}
	switch req.Agent {
	case "opencode":
		extractCmds = append(extractCmds,
			"mkdir -p /root/.local/share/opencode /root/.config/opencode/plugins /root/.config/opencode/skills",
			"cp /tmp/unpacked/opencode/*.json /root/.local/share/opencode/ 2>/dev/null; true",
			"[ -f /tmp/unpacked/opencode/opencode.jsonc ] && cp /tmp/unpacked/opencode/opencode.jsonc /root/.config/opencode/; true",
			"[ -d /tmp/unpacked/opencode/plugins ] && cp -r /tmp/unpacked/opencode/plugins/* /root/.config/opencode/plugins/ 2>/dev/null; true",
			"[ -d /tmp/unpacked/opencode/skills ] && cp -r /tmp/unpacked/opencode/skills/* /root/.config/opencode/skills/ 2>/dev/null; true",
		)
	case "pi":
		extractCmds = append(extractCmds,
			"mkdir -p /root/.pi/agent",
			"cp /tmp/unpacked/pi/* /root/.pi/agent/ 2>/dev/null; true",
		)
	case "crush":
		extractCmds = append(extractCmds,
			"mkdir -p /root/.config/crush /root/.local/share/crush",
			"[ -f /tmp/unpacked/crush/config/crush.json ] && mkdir -p /root/.config/crush && cp /tmp/unpacked/crush/config/crush.json /root/.config/crush/; true",
			"[ -f /tmp/unpacked/crush/data/crush.json ] && mkdir -p /root/.local/share/crush && cp /tmp/unpacked/crush/data/crush.json /root/.local/share/crush/; true",
		)
	}
	extractScript := strings.Join(extractCmds, " && ")
	if _, err := podman.ExecOutput(containerName, "sh", "-c", extractScript); err != nil {
		return nil, fmt.Errorf("extract configs in container: %w", err)
	}

	t.Logf("[9/9] running query (timeout=%ds): %s", req.TimeoutSeconds, req.Query)
	timeoutCmd := fmt.Sprintf("timeout %d %s", req.TimeoutSeconds, req.Query)
	output, execErr := podman.ExecOutput(containerName, "sh", "-c", timeoutCmd)
	t.Logf("[result] output=%q err=%v", output, execErr)
	return &Response{Output: output}, execErr
}

func configsExist(agent, homeDir string) bool {
	switch agent {
	case "opencode":
		dataDir := filepath.Join(homeDir, ".local", "share", "opencode")
		if entries, err := os.ReadDir(dataDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					return true
				}
			}
		}
		cfgPath := filepath.Join(homeDir, ".config", "opencode", "opencode.jsonc")
		if _, err := os.Stat(cfgPath); err == nil {
			return true
		}
		pluginsDir := filepath.Join(homeDir, ".config", "opencode", "plugins")
		if entries, err := os.ReadDir(pluginsDir); err == nil && len(entries) > 0 {
			return true
		}
		skillsDir := filepath.Join(homeDir, ".config", "opencode", "skills")
		if entries, err := os.ReadDir(skillsDir); err == nil && len(entries) > 0 {
			return true
		}
		return false
	case "pi":
		agentDir := filepath.Join(homeDir, ".pi", "agent")
		for _, f := range []string{"auth.json", "settings.json", "models.json"} {
			if _, err := os.Stat(filepath.Join(agentDir, f)); err == nil {
				return true
			}
		}
		return false
	case "crush":
		if _, err := os.Stat(filepath.Join(homeDir, ".config", "crush", "crush.json")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(homeDir, ".local", "share", "crush", "crush.json")); err == nil {
			return true
		}
		return false
	}
	return false
}
```
