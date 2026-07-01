package agentprovider

import (
	"fmt"
	"strings"

	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	claudeagent "github.com/xhd2015/agent-pro/agent/cli/claude"
	crushagent "github.com/xhd2015/agent-pro/agent/cli/crush"
	cursoragent "github.com/xhd2015/agent-pro/agent/cli/cursor"
	grokagent "github.com/xhd2015/agent-pro/agent/cli/grok"
	opencodeagent "github.com/xhd2015/agent-pro/agent/cli/opencode"
	piagent "github.com/xhd2015/agent-pro/agent/cli/pi"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

func Build(runnerID registry.AgentRunnerID, settingsPath, workspace string, env *exec.Env) (registry.AgentRunner, error) {
	id := registry.AgentRunnerID(strings.TrimSpace(string(runnerID)))
	if id == "" {
		return registry.AgentRunner{}, fmt.Errorf("agent runner id is required")
	}
	switch id {
	case registry.AgentRunnerCursor:
		cursorPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.CursorCLIPathSettingKey, registry.EnvCursorCLIPath, "", func() (string, error) {
			return cursoragent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("cursor-agent not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerCursor,
			Name: "Cursor",
			Agent: &cursoragent.CursorAgent{
				AgentPath:    cursorPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case registry.AgentRunnerCodex:
		codexPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.CodexCLIPathSettingKey, registry.EnvCodexCLIPath, "", func() (string, error) {
			return codexagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("codex not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerCodex,
			Name: "Codex",
			Agent: &codexagent.CodexAgent{
				AgentPath:    codexPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case registry.AgentRunnerOpencode:
		opencodePath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.OpencodeCLIPathSettingKey, registry.EnvOpencodeCLIPath, "", func() (string, error) {
			return opencodeagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("opencode not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerOpencode,
			Name: "Opencode",
			Agent: &opencodeagent.OpencodeAgent{
				AgentPath:    opencodePath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case registry.AgentRunnerFakeCodex:
		fakeCodexPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.FakeCodexCLIPathSettingKey, registry.EnvFakeCodexCLIPath, "", func() (string, error) {
			return codexagent.FindFakeCodexPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("fake-codex not found: %w (build it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerFakeCodex,
			Name: "Fake Codex",
			Agent: &codexagent.CodexAgent{
				AgentPath:    fakeCodexPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case registry.AgentRunnerCrush:
		crushPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.CrushCLIPathSettingKey, registry.EnvCrushCLIPath, "", func() (string, error) {
			return crushagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("crush not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerCrush,
			Name: "Crush",
			Agent: &crushagent.CrushAgent{
				AgentPath:    crushPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case registry.AgentRunnerPi:
		piPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.PiCLIPathSettingKey, registry.EnvPiCLIPath, "", func() (string, error) {
			return piagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("pi not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerPi,
			Name: "Pi",
			Agent: &piagent.PiAgent{
				AgentPath:    piPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case registry.AgentRunnerGrok:
		grokPath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.GrokCLIPathSettingKey, registry.EnvGrokCLIPath, "", func() (string, error) {
			return grokagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("grok not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerGrok,
			Name: "Grok",
			Agent: &grokagent.GrokAgent{
				AgentPath:    grokPath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	case registry.AgentRunnerClaude:
		claudePath, err := registry.ResolveConfiguredCLIPath(settingsPath, registry.ClaudeCLIPathSettingKey, registry.EnvClaudeCLIPath, "", func() (string, error) {
			return claudeagent.FindAgentPath(env)
		})
		if err != nil {
			return registry.AgentRunner{}, fmt.Errorf("claude not found: %w (install it or add it to PATH)", err)
		}
		return registry.AgentRunner{
			ID:   registry.AgentRunnerClaude,
			Name: "Claude",
			Agent: &claudeagent.ClaudeAgent{
				AgentPath:    claudePath,
				SettingsPath: settingsPath,
				Workspace:    workspace,
				Env:          env,
			},
		}, nil
	default:
		return registry.AgentRunner{}, fmt.Errorf("unknown agent runner id: %s", id)
	}
}
