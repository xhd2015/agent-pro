# Scenario

**Feature**: FormatTree renders ProcNode lists with tree-cli connectors

```
# pure formatter — no Resolve, no Lsof
nodes (PPID links) + TreeFormatOptions{ASCII}
  -> multi-line string: root line, then children with ├──/└── or +--/`--
doctest <- TreeText
```

## Preconditions

- Leaves set `Mode=format_tree`, `TreeNodes` (3-node agent-run chain), and
  `ASCII` true/false.
- Nodes form a single root with PPID outside the set (or PPID=1); children
  reference parent PIDs.
- Output lines are `connectors + PID + " " + Cmd` (Role not required on the
  printed line).

## Steps

1. Grouping tags Mode and installs the shared 3-node fixture (200→201→202).
2. Leaf sets ASCII true or false.
3. Run calls `FormatTree`; Assert checks connectors and pid/cmd presence.

## Context

- Shared fixture mirrors resolve-hit/grok/agent-run-tree topology (cmds only):
  - 200 PPID=1 agent-run run
  - 201 PPID=200 agent-run serve
  - 202 PPID=201 grok
- Connector style for a sole-child chain is locked by leaf ASSERT templates so
  implementers match one deterministic drawing (includes `├──`, `│`, `└──` in
  Unicode mode).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "format_tree"
	// Shared 3-node chain; leaves only flip ASCII.
	req.TreeNodes = []FixtureNode{
		{PID: 200, PPID: 1, Role: "input", Cmd: "/usr/local/bin/agent-run run --session-id=ignored-cli"},
		{PID: 201, PPID: 200, Role: "agent-run-serve", Cmd: "/usr/local/bin/agent-run serve --session-id=ignored-cli"},
		{PID: 202, PPID: 201, Role: "grok", Cmd: "/usr/local/bin/grok"},
	}
	return nil
}
```
