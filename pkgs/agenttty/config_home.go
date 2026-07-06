package agenttty

import (
	"os"
	"path/filepath"
	"strings"
)

const EnvAgentRunnerConfigHome = "AGENT_RUNNER_CONFIG_HOME"

// ResolveAgentRunnerConfigHome returns the flag value or AGENT_RUNNER_CONFIG_HOME env.
func ResolveAgentRunnerConfigHome(flagValue string) string {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(EnvAgentRunnerConfigHome))
}

// GrokHomeForRunner returns the grok data directory for session discovery.
func GrokHomeForRunner(configHome string) string {
	if v := ResolveAgentRunnerConfigHome(configHome); v != "" {
		return v
	}
	return GrokHome()
}

// CodexHomeForRunner returns the codex data directory for session discovery.
func CodexHomeForRunner(configHome string) string {
	if v := ResolveAgentRunnerConfigHome(configHome); v != "" {
		return v
	}
	return CodexHome()
}

// PrependCommandEnv prefixes argv with env(1) assignments for the PTY child process.
func PrependCommandEnv(argv []string, runnerID, configHome string) []string {
	env := RunnerConfigHomeEnv(runnerID, configHome)
	if len(env) == 0 {
		return argv
	}
	out := make([]string, 0, 1+len(env)+len(argv))
	out = append(out, "env")
	out = append(out, env...)
	out = append(out, argv...)
	return out
}

// RunnerConfigHomeEnv returns runner-specific env assignments for a PTY child.
func RunnerConfigHomeEnv(runnerID, configHome string) []string {
	configHome = strings.TrimSpace(configHome)
	if configHome == "" {
		return nil
	}
	switch strings.TrimSpace(runnerID) {
	case "codex-tty":
		return []string{"CODEX_HOME=" + configHome}
	default:
		return []string{envGrokHome + "=" + configHome}
	}
}

// AutoProvisionGrokConfigHome creates a temp grok home when using llm-mock-run-grok
// without an explicit config home so agent-run discovery matches the child process.
func AutoProvisionGrokConfigHome(runnerID, agentRunnerBinary, configHome string) (string, error) {
	if ResolveAgentRunnerConfigHome(configHome) != "" {
		return "", nil
	}
	if strings.TrimSpace(runnerID) != "grok-tty" {
		return "", nil
	}
	if !isLLMMockRunGrokBinary(agentRunnerBinary) {
		return "", nil
	}
	dir, err := os.MkdirTemp("", "agent-run-grok-home-*")
	if err != nil {
		return "", err
	}
	return dir, nil
}

func isLLMMockRunGrokBinary(agentRunnerBinary string) bool {
	spec := strings.TrimSpace(agentRunnerBinary)
	if spec == "" {
		return false
	}
	words, err := parseShellWords(spec)
	if err != nil || len(words) == 0 {
		return false
	}
	return filepath.Base(words[0]) == "llm-mock-run-grok"
}