package subagent

import "context"

// TestExported_runAgent wraps the unexported runAgent for doctest access.
func TestExported_runAgent(ctx context.Context, agentRunner, model, prompt, sessionID string, rawLog *sessionLogWriter) (string, error) {
	return runAgent(ctx, agentRunner, model, prompt, sessionID, rawLog, nil)
}

// TestExported_NewSessionLogWriter creates a sessionLogWriter for testing.
func TestExported_NewSessionLogWriter() *sessionLogWriter {
	return &sessionLogWriter{}
}
