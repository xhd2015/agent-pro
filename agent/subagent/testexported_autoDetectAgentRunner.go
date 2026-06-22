package subagent

// TestExported_autoDetectAgentRunner wraps the unexported autoDetectAgentRunner
// for doctest access.
func TestExported_autoDetectAgentRunner(c Config) (string, bool) {
	return AutoDetectAgentRunner(c)
}

// TestProcessNameFunc is a test hook that, when non-nil, overrides the
// PID → process-name lookup used by autoDetectAgentRunner during Priority 4
// (parent-process) detection. Tests set this to inject a fake process tree.
// The production getProcessName delegates to this hook before falling back
// to OS calls (ps on darwin, /proc on linux).
var TestProcessNameFunc func(pid int) string
