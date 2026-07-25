# Scenario

**Feature**: agent-run → serve → node wrap (no files) → codex leaf with rollout path

```
pid 300 agent-run run …          (input)
  -> 301 agent-run serve …
       -> 302 /usr/bin/node wrap.js   (other; empty Lsof)
            -> 303 /usr/local/bin/codex
                 Lsof(303) -> …/rollout-…-a1b2….jsonl
ResolveFromPID(300) -> Kind=codex, RunnerPID=303, Source=open-files+tree
```

## Preconditions

- Intermediate `node` is not a session runner (role `other`); empty open files.
- Prefer deeper leaf: only codex has session files → wins.

## Steps

1. Set `PID=300`.
2. Install four procs 300→301→302→303.
3. OpenFiles only for 303 with `codexRolloutPath(fixtureCodexSessionID)`.

## Context

- Do not treat node as runner even if it might open unrelated files (empty here).
- Source = `open-files+tree`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = 300
	req.Procs = []FixtureProc{
		{PID: 300, PPID: 1, Cmd: "/usr/local/bin/agent-run run --session-id=ignored-cli"},
		{PID: 301, PPID: 300, Cmd: "/usr/local/bin/agent-run serve --session-id=ignored-cli"},
		{PID: 302, PPID: 301, Cmd: "/usr/bin/node /app/wrap.js"},
		{PID: 303, PPID: 302, Cmd: "/usr/local/bin/codex"},
	}
	req.OpenFiles = map[int][]string{
		303: {codexRolloutPath(fixtureCodexSessionID)},
	}
	return nil
}
```
