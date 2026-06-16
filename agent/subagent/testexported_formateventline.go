package subagent

// TestExported_formatEventLine wraps the unexported formatEventLine for
// doctest access.
func TestExported_formatEventLine(line string) string {
	return formatEventLine(line)
}
