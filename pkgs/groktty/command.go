package groktty

import (
	"fmt"
	"os"
	"strings"

	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	grokagent "github.com/xhd2015/agent-pro/agent/cli/grok"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

const envGrokTTYCommand = "AGENT_RUN_GROK_TTY_COMMAND"
const envCodexTTYCommand = "AGENT_RUN_CODEX_TTY_COMMAND"

// BuildCommandArgv returns argv for the interactive grok TUI inside the PTY.
func BuildCommandArgv(env *exec.Env, settingsPath, agentPath, model, resumeSession string) ([]string, error) {
	return BuildGrokCommandArgv(env, settingsPath, agentPath, model, resumeSession)
}

// BuildGrokCommandArgv returns argv for the interactive grok TUI inside the PTY.
func BuildGrokCommandArgv(env *exec.Env, settingsPath, agentPath, model, resumeSession string) ([]string, error) {
	if hook := strings.TrimSpace(os.Getenv(envGrokTTYCommand)); hook != "" {
		return parseShellWords(hook)
	}

	agent := &grokagent.GrokAgent{
		AgentPath:    agentPath,
		SettingsPath: settingsPath,
		Env:          env,
	}
	path, err := resolveGrokPath(agent)
	if err != nil {
		return nil, err
	}

	args := []string{path, "--always-approve", "--permission-mode=bypassPermissions"}
	if model != "" {
		args = append(args, "--model", model)
	}
	if resumeSession != "" {
		args = append(args, "--resume", resumeSession)
	}
	return args, nil
}

// BuildCodexCommandArgv returns argv for the interactive Codex TUI inside the PTY.
func BuildCodexCommandArgv(env *exec.Env, settingsPath, agentPath, model, resumeSession string) ([]string, error) {
	if hook := strings.TrimSpace(os.Getenv(envCodexTTYCommand)); hook != "" {
		return parseShellWords(hook)
	}

	agent := &codexagent.CodexAgent{
		AgentPath:    agentPath,
		SettingsPath: settingsPath,
		Env:          env,
	}
	path, err := resolveCodexPath(agent)
	if err != nil {
		return nil, err
	}

	args := []string{path, "--dangerously-bypass-approvals-and-sandbox"}
	if model != "" {
		args = append(args, "--model", model)
	}
	if resumeSession != "" {
		args = append(args, "resume", resumeSession)
	}
	return args, nil
}

func resolveGrokPath(agent *grokagent.GrokAgent) (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		agent.SettingsPath,
		registry.GrokCLIPathSettingKey,
		registry.EnvGrokCLIPath,
		agent.AgentPath,
		func() (string, error) { return grokagent.FindAgentPath(agent.Env) },
	)
	if err != nil {
		return "", err
	}
	return path, nil
}

func resolveCodexPath(agent *codexagent.CodexAgent) (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		agent.SettingsPath,
		registry.CodexCLIPathSettingKey,
		registry.EnvCodexCLIPath,
		agent.AgentPath,
		func() (string, error) { return codexagent.FindAgentPath(agent.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("codex not found: %w", err)
	}
	return path, nil
}

// ParseShellWords splits a shell-style command line into argv words.
func ParseShellWords(line string) ([]string, error) {
	return parseShellWords(line)
}

func parseShellWords(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty command")
	}
	var words []string
	var cur strings.Builder
	var quote rune
	escaped := false
	for _, r := range line {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		switch {
		case r == '\\' && quote == 0:
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t':
			if cur.Len() > 0 {
				words = append(words, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote in %q", line)
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return words, nil
}
