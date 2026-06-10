# Agent Hub CLI

`agent-hub` is the command-line client for a local `agent-hubd` daemon. Agent
runners, hook adapters, dashboards, and workers use the CLI to publish and
consume normalized agent events without writing queue files directly.

## Design Goals

- Keep all queue mutation inside `agent-hubd`.
- Support producer calls from runner hooks with short, deterministic commands.
- Support Kafka-like consumers with independent cursors.
- Support batch fetches with `--limit`.
- Keep test integrations isolated through `cmd/fake-codex` and
  `cmd/fake-opencode`.

## Environment

```sh
AGENT_HUB_HOME=/path/to/.agent-hub
AGENT_HUB_SOCKET=/path/to/.agent-hub/agent-hub.sock
```

If `AGENT_HUB_HOME` is not set, the default is `.agent-hub` under the current
working directory. Tests must always set `AGENT_HUB_HOME` to a temporary
directory.

## Daemon Commands

```sh
agent-hub daemon start [--home DIR] [--socket PATH] [--http 127.0.0.1:0]
agent-hub daemon stop [--home DIR]
agent-hub daemon status [--home DIR] [--json]
```

`agent-hubd` owns the event log, consumer cursors, materialized session state,
and orchestration hooks. The CLI should fail clearly when the daemon is not
running:

```text
agent-hubd is not running; start it with: agent-hub daemon start
```

`--auto-start-daemon` can be added later, but v1 should avoid hidden background
process management.

## Producer Commands

### notify

```sh
agent-hub notify --json '{"event_type":"agent.session.started","runner":"codex"}'
agent-hub notify --file event.json
```

`notify` sends one normalized event to `agent-hubd`. The daemon assigns
`event_id`, `partition`, `offset`, and `received_at`.

Input:

```json
{
  "event_type": "agent.session.started",
  "runner": "codex",
  "runner_session_id": "codex-thread-123",
  "workspace": "/repo",
  "model": "gpt-5",
  "prompt": "initial user prompt",
  "occurred_at": "2026-06-10T12:00:00Z",
  "payload": {}
}
```

Output:

```json
{
  "event_id": "01J00000000000000000000000",
  "partition": "2026-06-10",
  "offset": 0,
  "received_at": "2026-06-10T12:00:01Z"
}
```

### hook notify

```sh
agent-hub hook notify --runner codex --event SessionStart
agent-hub hook notify --runner codex --event UserPromptSubmit
agent-hub hook notify --runner opencode --event session.created
```

`hook notify` reads native hook payload JSON from stdin, normalizes the payload,
and publishes the resulting event to the daemon. It is the preferred command for
runner hooks and fake-runner tests.

Example:

```sh
agent-hub hook notify --runner codex --event SessionStart < payload.json
```

The command should have a short default timeout. Hook failures should be visible
in stderr but should not corrupt queue files because only the daemon mutates
storage.

## Consumer Commands

### fetch

```sh
agent-hub fetch --consumer-id dashboard
agent-hub fetch --consumer-id dashboard --limit 10
agent-hub fetch --consumer-id dashboard --limit 10 --peek
```

`fetch` reads up to `--limit` events from the consumer's cursor. The default
limit is `1`. `--limit 0` is invalid. The daemon should enforce a configured
maximum, such as `1000`.

Default `fetch` advances the stored consumer cursor atomically to the position
after the returned batch. `--peek` returns the same batch shape without advancing
the cursor.

Output:

```json
{
  "consumer_id": "dashboard",
  "events": [
    {
      "event_id": "01J00000000000000000000000",
      "partition": "2026-06-10",
      "offset": 0,
      "event_type": "agent.session.started",
      "runner": "codex",
      "runner_session_id": "codex-thread-123",
      "received_at": "2026-06-10T12:00:01Z",
      "event": {}
    }
  ],
  "previous_cursor": {
    "partition": "2026-06-10",
    "offset": 0
  },
  "next_cursor": {
    "partition": "2026-06-10",
    "offset": 1
  },
  "has_more": false
}
```

`has_more` means the daemon can currently see additional events after
`next_cursor`. It does not predict future events.

### commit

```sh
agent-hub commit --consumer-id worker --cursor 2026-06-10:42
```

`commit` is reserved for a future `fetch --lease` or `fetch --no-advance` worker
mode. Plain v1 `fetch` is read-and-advance.

### replay

```sh
agent-hub replay --consumer-id dashboard --from 2026-06-01:0
```

`replay` resets a consumer cursor to a specific partition offset. It should
require an explicit cursor and should print the previous cursor.

## Inspection Commands

```sh
agent-hub status --json
agent-hub consumers --json
agent-hub sessions --json
agent-hub partitions --json
```

`sessions` reads materialized state maintained by the daemon. The event log
remains the source of truth.

## Fake Runner Commands

Fake runners must not read or mutate host runner configuration. They only use
explicit mock config files and temporary environment roots.

```sh
fake-codex exec --json --mock-config /tmp/mock.json "prompt"
fake-opencode run --format json --mock-config /tmp/mock.json "prompt"
fake-opencode models
```

Mock config files describe stdout events, hook events, hook commands, exit code,
stderr, delays, session IDs, and model metadata. `fake-codex --script` and
`FAKE_CODEX_SCRIPT` can remain as legacy aliases, but `--mock-config` is the
canonical integration-test surface.

## Exit Codes

- `0`: command succeeded
- `1`: invalid arguments, invalid JSON, or daemon request failed
- `2`: daemon is not running
- `3`: normalization failed for a hook payload
- `4`: consumer cursor is invalid or points to a missing partition

