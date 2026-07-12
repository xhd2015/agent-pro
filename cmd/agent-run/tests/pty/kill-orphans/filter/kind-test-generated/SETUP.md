# Scenario

**Feature**: `--kind=test-generated` matches serves with `TestGenerated` in path/argv (any PPID)

```
spawn child: <tmpdir>/TestGeneratedCase/bin/agent-run __serve_*__ … (PPID ≠ 1)
agent-run pty kill-orphans --dry-run --kind=test-generated --exe <tg-bin>
  -> stdout lists child PID
follow-up: agent-run pty kill-orphans --dry-run --exe <tg-bin>
  -> default does not list child (PPID ≠ 1, no --kind/--all)
```

## Preconditions

- Binary installed under a path segment containing `TestGenerated`.
- Child spawn (not double-fork) so default PPID1 filter excludes it.
- Follow-up CLI contrasts default selection in the same leaf.

## Steps

1. Prepare TestGenerated-path binary; spawn non-orphan serve with it.
2. Dry-run with `--kind=test-generated --exe <tg-bin>`.
3. Follow-up default dry-run with same `--exe`.
4. Assert kind lists PID; default does not; trailing `\n`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "kill-orphans"
	tgBin := serveBinaryForMarker(t, req, "test-generated")
	req.SpawnPlan = []ServeSpawnSpec{
		{
			Label:      "tg",
			Orphan:     false,
			PathMarker: "test-generated",
			SessionID:  "pty-kind-tg",
		},
	}
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--kind=test-generated",
		"--exe", tgBin,
	}
	req.FollowUpArgs = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--exe", tgBin,
	}
	return nil
}
```
