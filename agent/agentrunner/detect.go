// Package agentrunner provides lean parent-agent detection with no heavy deps.
// Consumers (e.g. tsk status auto-format) can import this package without
// pulling in subagent, logging, or UI code.
package agentrunner

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Options configures detection.
type Options struct {
	// AgentRunnerEnv is the env var name for explicit override (empty = skip).
	AgentRunnerEnv string
}

// TestProcessNameFunc, when non-nil, overrides PID → process-name lookup used
// during parent-process detection. Tests set this to inject a fake process tree.
var TestProcessNameFunc func(pid int) string

// Detect reports the parent agent runner if one is found.
// Priority:
//  1. Options.AgentRunnerEnv env override (if name non-empty and set)
//  2. CODEX_THREAD_ID → "codex"
//  3. PI_CODING_AGENT → "pi"
//  4. parent process name: opencode|pi|crush|codex|grok
//  5. grandparent: pi|grok
func Detect(opts Options) (runner string, ok bool) {
	if opts.AgentRunnerEnv != "" {
		if v := os.Getenv(opts.AgentRunnerEnv); v != "" {
			return strings.TrimSpace(v), true
		}
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
