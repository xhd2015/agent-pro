# Scenario

**Feature**: non-open headless — new-session prompt on argv only (no PTY re-inject);
banner hard-wait still applies when inject-ready is needed

```
# new-session normal submit: argv only (same inject policy as --open)
agent-run run --agent-runner grok-tty "once-only"
  + banner probe records ARGV + timed stdin
  -> PROMPT_ARG=once-only; STDIN_COUNT=0

# delayed banner: still no re-inject for new-session
agent-run run --agent-runner grok-tty "hi"
  + delayed GROK_TTY_BANNER
  -> wait banner → no PTY re-inject → exit 0
```

## Preconditions

- No `--open`; no INSTANT attach env (irrelevant for non-open attach).
- New-session `!NoSubmit`: do not re-inject (fix for real Grok double-submit).
- Resume / NoSubmit still inject (covered under open / resume trees, not this group).

## Steps

1. Grouping sets grok-tty non-open base args.
2. Leaves install banner argv/stdin probe fixtures + prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	clearOpenInstantAttach(req)
	req.Runner = "grok-tty"
	req.Args = []string{"run", "--agent-runner", "grok-tty"}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
