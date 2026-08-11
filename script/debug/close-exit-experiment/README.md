# open close→exit (production policy)

When the user closes the iTerm window for an `--open` session, treat it like
session end: reap the PTY agent and shut down `__serve__` (free resources).
The agent-run session meta remains so a later re-tag can **resume**.

## Policy

| Mode | Window red-close | Notes |
|------|------------------|--------|
| `--open` (default) | stopChild + serve exit | `OpenCloseExits()` default **on** |
| `--open` + `--keep-tty` | stopChild; serve may keep-alive | explicit keep |
| `--detach` | N/A (no window owner) | daemon stays |
| Ctrl-] detach | child kept (`detach_keep`) | intentional |

Opt out (old ghost-keep-alive open behavior):

```bash
export AGENT_RUN_OPEN_CLOSE_EXITS=0
```

## Mechanism

1. `AttachWriter(..., "screen")` → ptywrap **writer** role  
   bare WS disconnect without `detach_keep` → `stopChild()`
2. `--open` does not force `KeepTerminalAlive` when `OpenCloseExits()`  
   → after child exits, `__serve__` shuts down (non-keep-alive path)

Code: `pkgs/agenttty/open_close_exits.go`, `pkgs/agenttty/run.go`,  
`pkgs/agentui/run.go`, `pkgs/agentruncli/run_cmd.go`, `pkgs/agentrunapi/production.go`.

## Manual check after reinstall

```bash
# open a session, red-close the window, then:
spl agent-run status <session-id>
# expect: process dead, terminal missing/unreachable, exited true (if bound)
```
