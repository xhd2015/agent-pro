# Scenario

**Feature**: Run (start + wait-until-done) and RunJSON (schema + temp result file)

```
Run(prompt, dir, OpenTerminal?, StoreHome?, lifetime flags)
  -> launch detach | open-terminal follow-up
  -> wait until done
  -> optional /exit

RunJSON(prompt, schema)
  -> temp result path + schema appended
  -> Run
  -> read JSON string
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` exports `Run`,
  `RunJSON`, `RunOpts`, `RunJSONOpts`.
- **No real agent-run binary, iTerm, or grok.** Leaves inject Launch / Wait /
  OpenFn / SoftExit.
- Nested tree: `d.DOCTEST_ROOT` is `tests/agentrunapi/run`; module root is
  `../../..`.
- Parent P1 `tests/agentrunapi` leaves stay GREEN independently.
- Harness stores API errors on `Response.ErrString`.

## Steps

1. Root `Setup` seeds a default prompt.
2. Grouping `Setup` sets `req.Mode`.
3. Leaf fills flags; `Run` calls `agentrunapi.Run` / `RunJSON`.

## Context

- Default prompt `hello run`. WorkspaceDir is `t.TempDir()` when empty.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Prompt == "" {
		req.Prompt = "hello run"
	}
	return nil
}
```
