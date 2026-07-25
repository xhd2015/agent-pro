# Scenario

**Feature**: `proc resolve --json` prints kind and session id from fixture snapshot

```
AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT=<bare grok hit>
agent-pro proc resolve 100 --json
  -> exit 0
  -> stdout JSON includes kind=grok and session id
  -> no Unicode tree glyphs required in JSON body
```

## Preconditions

- Fixture matches library bare-input grok hit (pid 100, open session path).
- CLI must read `AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT` when set.

## Steps

1. Install Snapshot with pid 100 grok + open_files path containing fixture uuid.
2. Args: `proc resolve 100 --json`.
3. Assert exit 0, session id + kind present, no `├`/`└`/`│` required.

## Context

- JSON field names may be camelCase or snake_case (`session_id` / `SessionID` /
  `sessionId`); assert on substring presence of the uuid and `grok`.
- Stdout ends with trailing newline.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"proc", "resolve", "100", "--json"}
	req.Snapshot = &TestSnapshot{
		Procs: []SnapshotProc{
			{PID: 100, PPID: 1, Cmd: "/usr/local/bin/grok"},
		},
		OpenFiles: map[string][]string{
			"100": {
				"/tmp/fake-grok-home/.grok/sessions/2026-07/" + fixtureGrokSessionID + "/events.jsonl",
			},
		},
		GrokHome: "/tmp/fake-grok-home",
	}
	return nil
}
```
