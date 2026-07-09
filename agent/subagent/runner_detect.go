package subagent

import "github.com/xhd2015/agent-pro/agent/agentrunner"

// AutoDetectAgentRunner detects the parent agent runner for subagent execution.
// Explicit Config.AgentRunnerEnv wins, then well-known runner environment
// variables, then parent-process inspection.
//
// Implementation lives in the lean agentrunner package; this is a thin wrapper
// so existing callers and doctests keep working.
func AutoDetectAgentRunner(c Config) (runner string, detected bool) {
	// Bridge process-name test hook into agentrunner for detection doctests.
	prev := agentrunner.TestProcessNameFunc
	agentrunner.TestProcessNameFunc = TestProcessNameFunc
	defer func() { agentrunner.TestProcessNameFunc = prev }()

	return agentrunner.Detect(agentrunner.Options{
		AgentRunnerEnv: c.agentRunnerEnv(),
	})
}

func autoDetectAgentRunner(c Config) (runner string, detected bool) {
	return AutoDetectAgentRunner(c)
}
