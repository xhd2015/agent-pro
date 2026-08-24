package agenttty

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	grokagent "github.com/xhd2015/agent-pro/agent/cli/grok"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	codexargv "github.com/xhd2015/agent-pro/agent/codex/argv"
	"github.com/xhd2015/agent-pro/agent/exec"
)

const envGrokTTYCommand = "AGENT_RUN_GROK_TTY_COMMAND"
const envCodexTTYCommand = "AGENT_RUN_CODEX_TTY_COMMAND"
const envGrokHome = "GROK_HOME"

// BuildGrokCommandArgv returns argv for the interactive grok TUI inside the PTY.
func BuildGrokCommandArgv(env *exec.Env, settingsPath, agentRunnerBinary, model, resumeSession string) ([]string, error) {
	if hook := strings.TrimSpace(os.Getenv(envGrokTTYCommand)); hook != "" {
		return parseShellWords(hook)
	}

	path, userFlags, err := resolveGrokRunnerSpec(env, settingsPath, agentRunnerBinary)
	if err != nil {
		return nil, err
	}

	args := append([]string{path}, userFlags...)
	args = append(args, "--always-approve", "--permission-mode=bypassPermissions")
	if model != "" && !hasFlagPair(args, "--model") {
		args = append(args, "--model", model)
	}
	if resumeSession != "" && !hasFlagPair(args, "--resume") {
		args = append(args, "--resume", resumeSession)
	}
	return args, nil
}

// BuildCodexCommandArgv returns argv for the interactive Codex TUI inside the PTY.
// Delegates flag assembly to agent/codex/argv (Interactive preset).
func BuildCodexCommandArgv(env *exec.Env, settingsPath, agentRunnerBinary, model, resumeSession string) ([]string, error) {
	if hook := strings.TrimSpace(os.Getenv(envCodexTTYCommand)); hook != "" {
		// Hook overlay: update-check + hook-trust only (historical agent-run behavior).
		return codexargv.Argv(codexargv.Options{
			CommandOverride:    hook,
			BypassHookTrust:    true,
			DisableUpdateCheck: true,
		})
	}

	path, userFlags, err := resolveCodexRunnerSpec(env, settingsPath, agentRunnerBinary)
	if err != nil {
		return nil, err
	}

	opts := codexargv.Interactive()
	opts.Bin = path
	opts.UserFlags = userFlags
	opts.Model = model
	opts.ResumeSession = resumeSession
	return codexargv.Argv(opts)
}

// EnsureCodexConfigFlag ensures argv has paired "-c" "key=value" (preferred form).
func EnsureCodexConfigFlag(args []string, key, value string) []string {
	return codexargv.EnsureConfigFlag(args, key, value)
}

// ApplyCodexReasoningEffort: empty/whitespace effort → return args unchanged;
// otherwise EnsureCodexConfigFlag(args, "model_reasoning_effort", effort).
func ApplyCodexReasoningEffort(args []string, effort string) []string {
	effort = strings.TrimSpace(effort)
	if effort == "" {
		return args
	}
	return EnsureCodexConfigFlag(args, "model_reasoning_effort", effort)
}

// EnsureCodexBoolFlag ensures argv contains a bare boolean flag token exactly once.
func EnsureCodexBoolFlag(args []string, flag string) []string {
	return codexargv.EnsureBoolFlag(args, flag)
}

// hasCodexConfigKey reports whether args already contain -c/--config for key.
func hasCodexConfigKey(args []string, key string) bool {
	return codexargv.HasConfigKey(args, key)
}

func resolveGrokRunnerSpec(env *exec.Env, settingsPath, spec string) (path string, userFlags []string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		agent := &grokagent.GrokAgent{SettingsPath: settingsPath, Env: env}
		path, err = resolveGrokPath(agent)
		return path, nil, err
	}
	words, err := parseShellWords(spec)
	if err != nil {
		return "", nil, err
	}
	path, err = lookupRunnerBinary(env, words[0], func() (string, error) {
		agent := &grokagent.GrokAgent{SettingsPath: settingsPath, Env: env}
		return resolveGrokPath(agent)
	})
	if err != nil {
		return "", nil, err
	}
	return path, words[1:], nil
}

func resolveCodexRunnerSpec(env *exec.Env, settingsPath, spec string) (path string, userFlags []string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		agent := &codexagent.CodexAgent{SettingsPath: settingsPath, Env: env}
		path, err = resolveCodexPath(agent)
		return path, nil, err
	}
	words, err := parseShellWords(spec)
	if err != nil {
		return "", nil, err
	}
	path, err = lookupRunnerBinary(env, words[0], func() (string, error) {
		agent := &codexagent.CodexAgent{SettingsPath: settingsPath, Env: env}
		return resolveCodexPath(agent)
	})
	if err != nil {
		return "", nil, err
	}
	return path, words[1:], nil
}

func lookupRunnerBinary(env *exec.Env, binary string, fallback func() (string, error)) (string, error) {
	binary = strings.TrimSpace(binary)
	if binary == "" {
		if fallback == nil {
			return "", fmt.Errorf("runner binary is required")
		}
		return fallback()
	}
	if filepath.IsAbs(binary) {
		return binary, nil
	}
	if strings.Contains(binary, "/") || strings.Contains(binary, string(os.PathSeparator)) {
		return binary, nil
	}
	if env != nil {
		if path, err := env.LookPath(binary); err == nil {
			return path, nil
		}
	}
	return binary, nil
}

func hasFlagPair(args []string, flag string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			return true
		}
	}
	return false
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