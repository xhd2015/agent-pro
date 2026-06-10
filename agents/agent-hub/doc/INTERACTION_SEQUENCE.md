# Agent Hub Interaction Sequences

All diagrams are written in Mermaid and describe the expected v1 interaction
patterns.

## Daemon Startup

```mermaid
sequenceDiagram
  participant User
  participant CLI as agent-hub CLI
  participant Daemon as agent-hubd
  participant Store as File Storage

  User->>CLI: agent-hub daemon start
  CLI->>Daemon: spawn daemon
  Daemon->>Store: acquire storage lock
  Daemon->>Store: scan events/YYYY/MM/DD
  Daemon->>Store: rebuild missing indexes
  Daemon->>Store: rebuild session state
  Daemon->>Store: validate consumer cursors
  Daemon-->>CLI: ready
  CLI-->>User: daemon started
```

## Codex Hook Notification

```mermaid
sequenceDiagram
  participant Codex
  participant Hook as agent-hub hook notify
  participant Daemon as agent-hubd
  participant Store as Date-Partitioned Log

  Codex->>Hook: SessionStart payload on stdin
  Hook->>Hook: normalize Codex event
  Hook->>Daemon: notify(agent.session.started)
  Daemon->>Daemon: assign event_id, received_at, partition, offset
  Daemon->>Store: append events/YYYY/MM/DD/events.jsonl
  Daemon->>Store: update sessions/active
  Daemon-->>Hook: event_id, partition, offset
  Hook-->>Codex: exit 0
```

## Codex Prompt Correlation

```mermaid
sequenceDiagram
  participant Codex
  participant Hook as agent-hub hook notify
  participant Daemon as agent-hubd
  participant Session as Session Projector

  Codex->>Hook: UserPromptSubmit payload on stdin
  Hook->>Hook: normalize to agent.prompt.submitted
  Hook->>Daemon: notify(normalized event)
  Daemon->>Session: correlate runner_session_id or local correlation id
  Session->>Session: attach initial prompt if absent
  Daemon-->>Hook: accepted
```

## opencode Plugin Notification

```mermaid
sequenceDiagram
  participant Opencode as opencode
  participant Plugin as Agent Hub Plugin
  participant Hook as agent-hub hook notify
  participant Daemon as agent-hubd
  participant Store as Date-Partitioned Log

  Opencode->>Plugin: session.created
  Plugin->>Hook: execute hook notify with payload stdin
  Hook->>Hook: normalize opencode event
  Hook->>Daemon: notify(agent.session.started)
  Daemon->>Store: append event
  Daemon->>Store: update session state
  Daemon-->>Hook: accepted
  Hook-->>Plugin: exit 0
```

## Batch Fetch With Cursor Advance

```mermaid
sequenceDiagram
  participant Consumer
  participant CLI as agent-hub CLI
  participant Daemon as agent-hubd
  participant Cursor as consumers/dashboard.cursor.json
  participant Log as Event Partitions

  Consumer->>CLI: agent-hub fetch --consumer-id dashboard --limit 10
  CLI->>Daemon: fetch(dashboard, limit=10, peek=false)
  Daemon->>Cursor: load current cursor
  Daemon->>Log: read up to 10 events from cursor
  Log-->>Daemon: events, next_cursor, has_more
  Daemon->>Cursor: save next_cursor
  Daemon-->>CLI: batch response
  CLI-->>Consumer: JSON events
```

## Peek Fetch Without Cursor Advance

```mermaid
sequenceDiagram
  participant Consumer
  participant CLI as agent-hub CLI
  participant Daemon as agent-hubd
  participant Cursor as Consumer Cursor
  participant Log as Event Partitions

  Consumer->>CLI: agent-hub fetch --consumer-id dashboard --limit 10 --peek
  CLI->>Daemon: fetch(dashboard, limit=10, peek=true)
  Daemon->>Cursor: load current cursor
  Daemon->>Log: read up to 10 events
  Log-->>Daemon: events, next_cursor
  Daemon-->>CLI: batch response
  CLI-->>Consumer: JSON events
  Note over Cursor: Cursor is unchanged
```

## Crossing Date Partitions

```mermaid
sequenceDiagram
  participant Daemon as agent-hubd
  participant Day1 as events/2026/06/10/events.jsonl
  participant Day2 as events/2026/06/11/events.jsonl
  participant Cursor as Consumer Cursor

  Daemon->>Cursor: read cursor 2026-06-10:99
  Daemon->>Day1: read offset 99
  Day1-->>Daemon: event at 99, EOF after event
  Daemon->>Day2: read offset 0
  Day2-->>Daemon: event at 0
  Daemon->>Cursor: save 2026-06-11:1
```

## Fake Runner Integration Test

```mermaid
sequenceDiagram
  participant Test as Go Test
  participant Fake as fake-codex or fake-opencode
  participant Hook as agent-hub hook notify
  participant Daemon as agent-hubd
  participant Store as Temp AGENT_HUB_HOME

  Test->>Daemon: start with temp AGENT_HUB_HOME
  Test->>Fake: run --mock-config mock.json
  Fake->>Hook: fire configured hook payload
  Hook->>Daemon: notify normalized event
  Daemon->>Store: append date-partitioned event
  Fake-->>Test: stdout events and exit code
  Test->>Daemon: fetch --consumer-id test --limit 10
  Daemon-->>Test: stored hook events
```

## Daemon Shutdown

```mermaid
sequenceDiagram
  participant User
  participant CLI as agent-hub CLI
  participant Daemon as agent-hubd
  participant Store as File Storage

  User->>CLI: agent-hub daemon stop
  CLI->>Daemon: shutdown request
  Daemon->>Store: flush pending writes
  Daemon->>Store: write health stopped
  Daemon->>Store: release lock
  Daemon-->>CLI: stopped
  CLI-->>User: daemon stopped
```

