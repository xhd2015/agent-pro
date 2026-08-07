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
	return ApplyChildProcessEnv(argv, runnerID, configHome, nil, nil, false)
}

// ApplyChildProcessEnv prefixes argv with env(1) so the PTY child receives:
// - env entries (KEY=VALUE, later wins for the same KEY)
// - PATH with prependPaths joined ahead of the current PATH
// - GROK_HOME / CODEX_HOME from configHome
// - when color is true (applied last, wins over parent + -e):
//   - env -u NO_COLOR
//   - FORCE_COLOR=1 CLICOLOR=1 CLICOLOR_FORCE=1
//   - TERM=xterm-256color when effective TERM is empty or "dumb"
func ApplyChildProcessEnv(argv []string, runnerID, configHome string, prependPaths, envEntries []string, color bool) []string {
	assignments := make([]string, 0, len(envEntries)+8)
	// Effective TERM after parent + user -e (last-win); color policy may rewrite.
	effectiveTERM := os.Getenv("TERM")
	// Preserve order for meta/logging; env(1) last-wins for duplicate keys.
	for _, e := range envEntries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		key, val, hasEq := strings.Cut(e, "=")
		if hasEq && key == "TERM" {
			effectiveTERM = val
		}
		// Color policy wins over -e NO_COLOR=…: drop so -u can clear parent too.
		if color && hasEq && key == "NO_COLOR" {
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

	// Color force policy last: unset NO_COLOR, set force keys, TERM policy.
	// When effective TERM is empty/dumb → xterm-256color; otherwise pass through
	// the effective value so PTY defaults do not clobber a good parent TERM.
	var unset []string
	if color {
		unset = append(unset, "NO_COLOR")
		assignments = append(assignments,
			"FORCE_COLOR=1",
			"CLICOLOR=1",
			"CLICOLOR_FORCE=1",
		)
		if effectiveTERM == "" || effectiveTERM == "dumb" {
			assignments = append(assignments, "TERM=xterm-256color")
		} else {
			assignments = append(assignments, "TERM="+effectiveTERM)
		}
	}

	if len(assignments) == 0 && len(unset) == 0 {
		return argv
	}
	out := make([]string, 0, 1+2*len(unset)+len(assignments)+len(argv))
	out = append(out, "env")
	for _, k := range unset {
		out = append(out, "-u", k)
	}
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
