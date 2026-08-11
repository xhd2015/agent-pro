# close-exit-experiment (human debug)

Conceptual experiment: **iTerm red-close ≈ exit** for `--open` sessions.

## Mechanism (env gate)

```bash
export AGENT_RUN_OPEN_CLOSE_EXITS=1
```

| Knob | Baseline | Experiment |
|------|----------|------------|
| AttachWriter mode | `attach` (attacher — no stopChild) | `screen` (writer → stopChild on bare close) |
| KeepTerminalAlive for open | forced true | left false → serve exits after child dies |
| Detach | keep-alive daemon | **unchanged** |

Propagation path on red-close:

```text
iTerm close → AttachWriter process dies → WS drop (no detach_keep)
  → ptywrap writer close → stopChild()  # kill PTY agent
  → __serve__ Wait done → !KeepAlive → shutdown + remove registry
```

## Build

```bash
cd external/agent-pro-master-2026-08-11-1   # or agent-pro checkout
go build -o /tmp/agent-run-close-exit ./cmd/agent-run
```

## Run (simple long-lived child, no codex)

```bash
export AGENT_RUN_OPEN_CLOSE_EXITS=1
export AGENT_RUN_HOME=/tmp/close-exit-agent-run-home
mkdir -p /tmp/close-exit-ws "$AGENT_RUN_HOME"

SID=close-exit-exp-1
/tmp/agent-run-close-exit run --open \
  --session-id "$SID" \
  --agent-runner=commandcode-tty \
  --agent-runner-binary=/bin/bash \
  --dir /tmp/close-exit-ws \
  -- 'echo CLOSE_EXIT_EXP; sleep 3600'
```

Then:

```bash
/tmp/agent-run-close-exit status "$SID"
# close the iTerm window (red icon)
sleep 2
/tmp/agent-run-close-exit status "$SID"
pgrep -fl "close-exit-exp-1|__serve_.*close-exit" || echo "no serve"
```

## Expected after red-close (experiment ON)

- process serve: **dead / missing**
- `exited: true` or session not reachable
- no PPID=1 orphan `__serve__` for this session

## Expected after red-close (experiment OFF / stock)

- serve **alive**, PPID=1, sendable often yes
- `exited: false` → ModeSend ghost
