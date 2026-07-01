# Scenario

**Feature**: Two-turn Ask() resumes a prior claude session via --resume <id>

```
# turn 1: fresh query, capture LastSessionID from system/result
session-resume -> ClaudeAgent.Ask(prompt) -> claude -p <prompt> --output-format stream-json --verbose
ClaudeAgent <- claude (session_id)
# turn 2: resume reuses LastSessionID via --resume
session-resume -> ClaudeAgent.Ask(resumePrompt, SessionID=lastID)
ClaudeAgent -> claude -p <resumePrompt> --output-format stream-json --verbose --resume <lastID>
ClaudeAgent <- claude (assistant text, result)
```

## Preconditions
- The `claude` binary is available in PATH.
- Session resume runs two queries: first captures `LastSessionID`, second
  reuses it via `--resume`.

## Steps
1. Set the initial prompt to a deterministic question.
2. Set the resume prompt to ask what was previously asked.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "Reply with exactly the word: pong"
	req.ResumePrompt = "what did I ask you about?"
	return nil
}
```
