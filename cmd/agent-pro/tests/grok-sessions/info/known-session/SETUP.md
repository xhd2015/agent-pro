# Scenario

**Feature**: info returns full summary, file paths, and token usage

```
# rich summary.json + signals.json under encoded cwd
writeGrokSessionOpts -> sessions.Info -> FormatInfoText(now)

# output includes metadata, Files block, Tokens block
terminal key-value text
```

## Preconditions

- `signals.json` is present with non-zero token fields.
- `updates.jsonl` is absent (reported as missing).

## Steps

1. Write a session with model, agent, git, sandbox, and message counts.
2. Write `signals.json` with context token usage.
3. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const knownSessionID = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaaaa"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = knownSessionID
	writeGrokSessionOpts(t, req.GrokHome, knownSessionID,
		"2026-07-03T13:00:00.000Z",
		"/tmp/grok-known-project",
		"Refactor auth module",
		grokSessionOpts{
			NumMessages:      90,
			NumChatMessages:  40,
			CreatedAt:        "2026-07-03T10:00:00.000Z",
			UpdatedAt:        "2026-07-03T12:30:00.000Z",
			CurrentModelID:   "grok-composer-2.5-fast",
			AgentName:        "cursor",
			SandboxProfile:   "off",
			GitRootDir:       "/tmp/grok-known-project",
			HeadBranch:       "master-2026-07-03-1",
			HeadCommit:       "97433b50",
			WriteSignals:     true,
			ContextTokens:    75085,
			ContextWindow:    200000,
			ContextUsagePct:  38,
			TokensBeforeComp: 0,
		})
	return nil
}
```