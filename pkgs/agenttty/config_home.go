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
	return ApplyChildProcessEnv(argv, runnerID, configHome, nil, nil)
}

// ApplyChildProcessEnv prefixes argv with env(1) so the PTY child receives:
// - env entries (KEY=VALUE, later wins for the same KEY)
// - PATH with prependPaths joined ahead of the current PATH
// - GROK_HOME / CODEX_HOME from configHome
func ApplyChildProcessEnv(argv []string, runnerID, configHome string, prependPaths, envEntries []string) []string {
	assignments := make([]string, 0, len(envEntries)+2)
	// Preserve order for meta/logging; env(1) last-wins for duplicate keys.
	for _, e := range envEntries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		assignments = append(assignments, e)
	}
	if len(prependPaths) > 0 {
		parts := make([]string, 0, len(prependPaths)+1)
		for _, p := range prependPaths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			parts = append(parts, p)
		}
		if len(parts) > 0 {
			pathVal := strings.Join(parts, string(os.PathListSeparator))
			if cur := os.Getenv("PATH"); cur != "" {
				pathVal = pathVal + string(os.PathListSeparator) + cur
			}
			assignments = append(assignments, "PATH="+pathVal)
		}
	}
	assignments = append(assignments, RunnerConfigHomeEnv(runnerID, configHome)...)
	if len(assignments) == 0 {
		return argv
	}
	out := make([]string, 0, 1+len(assignments)+len(argv))
	out = append(out, "env")
	out = append(out, assignments...)
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
