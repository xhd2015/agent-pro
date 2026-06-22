package subagent

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// AutoDetectAgentRunner detects the parent agent runner for subagent execution.
// Explicit Config.AgentRunnerEnv wins, then well-known runner environment
// variables, then parent-process inspection.
func AutoDetectAgentRunner(c Config) (runner string, detected bool) {
	if v := os.Getenv(c.agentRunnerEnv()); v != "" {
		return strings.TrimSpace(v), true
	}
	if v := os.Getenv("CODEX_THREAD_ID"); v != "" {
		return "codex", true
	}
	if v := os.Getenv("PI_CODING_AGENT"); v != "" {
		return "pi", true
	}
	if ppid := os.Getppid(); ppid > 0 {
		comm := getProcessName(ppid)
		switch strings.ToLower(comm) {
		case "opencode":
			return "opencode", true
		case "pi":
			return "pi", true
		case "crush":
			return "crush", true
		case "codex":
			return "codex", true
		case "grok":
			return "grok", true
		}
		if pppid := getParentPid(ppid); pppid > 0 {
			switch strings.ToLower(getProcessName(pppid)) {
			case "pi":
				return "pi", true
			case "grok":
				return "grok", true
			}
		}
	}
	return "", false
}

func autoDetectAgentRunner(c Config) (runner string, detected bool) {
	return AutoDetectAgentRunner(c)
}

func getProcessName(pid int) string {
	if TestProcessNameFunc != nil {
		return TestProcessNameFunc(pid)
	}
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("ps", "-o", "comm=", "-p", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}
	return ""
}

func getParentPid(pid int) int {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("ps", "-o", "ppid=", "-p", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			return -1
		}
		var ppid int
		if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &ppid); err != nil {
			return -1
		}
		return ppid
	}
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			return -1
		}
		s := string(data)
		closeParen := strings.LastIndex(s, ")")
		if closeParen < 0 {
			return -1
		}
		fields := strings.Fields(strings.TrimSpace(s[closeParen+1:]))
		if len(fields) < 2 {
			return -1
		}
		var ppid int
		if _, err := fmt.Sscanf(fields[1], "%d", &ppid); err != nil {
			return -1
		}
		return ppid
	}
	return -1
}
