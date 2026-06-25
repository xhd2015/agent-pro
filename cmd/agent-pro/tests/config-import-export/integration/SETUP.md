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
