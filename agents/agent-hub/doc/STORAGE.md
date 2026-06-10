# Agent Hub Storage

`agent-hubd` uses files only. There is no SQLite, embedded database, or external
broker. The daemon is the only process that writes storage files.

## Storage Root

```text
.agent-hub/
  daemon.pid
  agent-hub.sock

  events/
    2026/
      05/
        30/
          events.jsonl
          events.idx
      06/
        01/
          events.jsonl
          events.idx

  consumers/
    dashboard.cursor.json
    metrics-worker.cursor.json

  sessions/
    active/
      codex-thread-123.json
    completed/
      codex-thread-123.json
    failed/
      codex-thread-999.json

  runtime/
    daemon-state.json
    health.json

  dead-letter/
    2026/
      06/
        10/
          invalid-events.jsonl
```

The event log is the source of truth. Consumer cursors and session files are
derived state that can be rebuilt or repaired from the event partitions.

## Date Partitions

Events are partitioned by daemon receive date:

```text
events/YYYY/MM/DD/events.jsonl
```

The partition is derived from `received_at`, not from producer-provided
timestamps such as `started_at` or `occurred_at`. This avoids clock-skewed or
late producers writing into old partitions unexpectedly.

Example:

```text
events/2026/06/10/events.jsonl
```

The logical partition name is:

```text
2026-06-10
```

## Event Envelope

Each JSONL row stores one daemon envelope:

```json
{
  "schema_version": "agent-hub.event.v1",
  "event_id": "01J00000000000000000000000",
  "partition": "2026-06-10",
  "offset": 42,
  "received_at": "2026-06-10T12:00:01Z",
  "producer": {
    "cli_version": "v0.1.0",
    "hostname": "devbox",
    "pid": 12345
  },
  "event": {
    "event_type": "agent.session.started",
    "runner": "codex",
    "runner_session_id": "codex-thread-123",
    "workspace": "/repo",
    "model": "gpt-5",
    "prompt": "initial user prompt",
    "occurred_at": "2026-06-10T12:00:00Z",
    "payload": {}
  }
}
```

`offset` is the zero-based event position inside one partition. It is logical
queue terminology, not a public reference to a file line.

## Consumer Cursor

Each consumer has an independent cursor:

```json
{
  "consumer_id": "dashboard",
  "cursor": {
    "partition": "2026-06-10",
    "offset": 43
  },
  "updated_at": "2026-06-10T12:05:00Z"
}
```

The cursor points to the next event to fetch. If the cursor points to EOF for a
partition, the daemon advances to the next available date partition.

## Fetch Across Partitions

`fetch --limit N` can return events from multiple partitions:

```json
{
  "events": [
    {
      "partition": "2026-06-10",
      "offset": 99
    },
    {
      "partition": "2026-06-11",
      "offset": 0
    }
  ],
  "next_cursor": {
    "partition": "2026-06-11",
    "offset": 1
  }
}
```

The daemon lists partitions chronologically by directory name. Missing days are
skipped.

## Index Files

`events.idx` is optional but recommended. It maps partition offsets to byte
positions inside `events.jsonl`:

```json
{"offset":0,"byte":0}
{"offset":1,"byte":381}
{"offset":2,"byte":762}
```

The log remains valid if the index is missing. The daemon can rebuild
`events.idx` by scanning `events.jsonl`.

## Materialized Sessions

Session state is maintained for efficient inspection and orchestration:

```json
{
  "runner_session_id": "codex-thread-123",
  "runner": "codex",
  "status": "active",
  "workspace": "/repo",
  "model": "gpt-5",
  "started_at": "2026-06-10T12:00:01Z",
  "finished_at": null,
  "last_event": {
    "partition": "2026-06-10",
    "offset": 42
  }
}
```

Sessions can span date partitions. A `session.started` event can be stored under
one date while `session.finished` is stored under the next date. Correlation is
by `runner` and `runner_session_id`.

## Locking

The daemon serializes mutations internally. If multiple daemon instances are
accidentally started with the same home, the active daemon must own an exclusive
lock under the storage root before accepting requests.

Only these operations mutate storage:

- append event envelope
- update partition index
- update consumer cursor
- update materialized session state
- write dead-letter event

CLI clients, hooks, and consumers do not write queue files directly.

## Dead Letters

Invalid normalized events are rejected before append. Invalid native hook
payloads can be recorded in daily dead-letter files:

```text
dead-letter/YYYY/MM/DD/invalid-events.jsonl
```

Dead-letter rows should include the runner, native event name, error, and
payload when safe.

## Recovery

On startup, the daemon should:

1. acquire the storage lock
2. scan partitions in chronological order
3. rebuild missing index files
4. rebuild in-memory session indexes
5. validate consumer cursors
6. expose health status

If materialized session files are missing or stale, they should be rebuilt from
the event log.

## Test Storage Rules

Tests must set:

```text
AGENT_HUB_HOME=<tempdir>/.agent-hub
```

Tests must not write to:

```text
~/.agent-hub
~/.codex
~/.config/opencode
```

