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

// ChildEnvSpec is the pure child-env policy for HeadlessRun:
// Set is KEY=VALUE for CommandEnv; Unset is bare keys for CommandUnset.
// Never compose argv with env(1).
type ChildEnvSpec struct {
	Set   []string
	Unset []string
}

// BuildChildProcessEnv builds pure Set/Unset policy for the PTY agent child.
// parentTERM is injectable (tests pass it; production may pass os.Getenv("TERM")).
// Policy matches the old ApplyChildProcessEnv env(1) composition without the
// env binary / argv mutation:
//   - env entries (KEY=VALUE; order preserved; last-wins at merge)
//   - PATH with prependPaths joined ahead of the current process PATH
//   - GROK_HOME / CODEX_HOME from configHome
//   - when color is true (applied last, wins over parent + -e):
//   - Unset NO_COLOR; drop user -e NO_COLOR=… from Set
//   - FORCE_COLOR=1 CLICOLOR=1 CLICOLOR_FORCE=1
//   - TERM=xterm-256color when effective TERM is empty or "dumb"
func BuildChildProcessEnv(runnerID, configHome string, prependPaths, envEntries []string, color bool, parentTERM string) ChildEnvSpec {
	set := make([]string, 0, len(envEntries)+8)
	// Effective TERM after parent + user -e (last-win); color policy may rewrite.
	effectiveTERM := parentTERM
	for _, e := range envEntries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		key, val, hasEq := strings.Cut(e, "=")
		if hasEq && key == "TERM" {
			effectiveTERM = val
		}
		// Color policy wins over -e NO_COLOR=…: drop so Unset can clear parent too.
		if color && hasEq && key == "NO_COLOR" {
			continue
		}
		set = append(set, e)
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
			set = append(set, "PATH="+pathVal)
		}
	}
	set = append(set, RunnerConfigHomeEnv(runnerID, configHome)...)

	var unset []string
	if color {
		unset = append(unset, "NO_COLOR")
		set = append(set,
			"FORCE_COLOR=1",
			"CLICOLOR=1",
			"CLICOLOR_FORCE=1",
		)
		if effectiveTERM == "" || effectiveTERM == "dumb" {
			set = append(set, "TERM=xterm-256color")
		} else {
			set = append(set, "TERM="+effectiveTERM)
		}
	}
	return ChildEnvSpec{Set: set, Unset: unset}
}

// PrependCommandEnv is a legacy no-op identity: child env is applied via
// BuildChildProcessEnv + HeadlessRun CommandEnv/CommandUnset, not env(1).
func PrependCommandEnv(argv []string, runnerID, configHome string) []string {
	return ApplyChildProcessEnv(argv, runnerID, configHome, nil, nil, false)
}

// ApplyChildProcessEnv is a transitional pure-argv helper. It no longer
// prefixes argv with env(1); policy belongs in BuildChildProcessEnv and is
// applied through HeadlessRun CommandEnv/CommandUnset. Argv is returned
// unchanged (callers keep pure agent command heads).
func ApplyChildProcessEnv(argv []string, runnerID, configHome string, prependPaths, envEntries []string, color bool) []string {
	_, _, _, _, _ = runnerID, configHome, prependPaths, envEntries, color
	return argv
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
