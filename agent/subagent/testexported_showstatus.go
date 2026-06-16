package subagent

// TestExported_showStatus wraps the unexported showStatus for doctest access.
func TestExported_showStatus(c Config, opts Options) error {
	return showStatus(c, opts)
}
