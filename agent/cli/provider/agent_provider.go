package agentprovider

import (
	"fmt"
	"strings"

	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	cursoragent "github.com/xhd2015/agent-pro/agent/cli/cursor"
	opencodeagent "github.com/xhd2015/agent-pro/agent/cli/opencode"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

func Build(runnerID, settingsPath, workspace string, env *exec.Env) (registry.AgentRunner, error) {
	id := strings.TrimSpace(runnerID)
	if id == "" {
		return registry.AgentRunner{}, fmt.Errorf("agent runner id is required")
	}
	switch id {
	case "cursor":
		cursorPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.CursorCLIPathSettingKey, "", func() (string, error) {
			return cursoragent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("cursor-agent not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   "cursor",
			Name: "Cursor",
			Agent: &cursoragent.CursorAgent{
				AgentPath:    cursorPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case "codex":
		codexPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.CodexCLIPathSettingKey, "", func() (string, error) {
			return codexagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("codex not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   "codex",
			Name: "Codex",
			Agent: &codexagent.CodexAgent{
				AgentPath:    codexPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case "opencode":
		opencodePath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.OpencodeCLIPathSettingKey, "", func() (string, error) {
			return opencodeagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("opencode not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   "opencode",
			Name: "Opencode",
			Agent: &opencodeagent.OpencodeAgent{
				AgentPath:    opencodePath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	default:
		return registry.AgentRunner{}, fmt.Errorf("unknown agent runner id: %s", id)
	}
}
