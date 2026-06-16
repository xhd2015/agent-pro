package subagent

// TestExported_traceSession wraps the unexported traceSession for doctest access.
func TestExported_traceSession(c Config, opts Options) error {
	return traceSession(c, opts)
}
