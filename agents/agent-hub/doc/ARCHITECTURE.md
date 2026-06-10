# Agent Hub Architecture

Agent Hub is a local event broker and orchestration daemon for agent runner
sessions. Producers publish normalized lifecycle events; consumers read from
independent cursors.

## Components

```mermaid
flowchart LR
  Codex[Codex Hooks] --> CodexAdapter[Codex Hook Adapter]
  Opencode[opencode Plugin Hooks] --> OpencodeAdapter[opencode Hook Adapter]
  FakeCodex[cmd/fake-codex] --> CodexAdapter
  FakeOpencode[cmd/fake-opencode] --> OpencodeAdapter

  CodexAdapter --> CLI[agent-hub CLI]
  OpencodeAdapter --> CLI
  CLI --> Socket[Unix Socket API]
  Socket --> Daemon[agent-hubd]

  Daemon --> Log[Date-Partitioned Event Log]
  Daemon --> Cursors[Consumer Cursors]
  Daemon --> Sessions[Materialized Sessions]

  Dashboard[Dashboard Consumer] --> CLI
  Worker[Worker Consumer] --> CLI
  CLI --> Socket
```

## Daemon Responsibilities

`agent-hubd` is the only process that mutates storage. It is responsible for:

- validating normalized events
- assigning `event_id`, `partition`, `offset`, and `received_at`
- appending to date-partitioned JSONL logs
- updating consumer cursors atomically
- maintaining materialized session state
- rebuilding indexes on startup
- exposing health and inspection data

## Producer Flow

```mermaid
flowchart TD
  Runner[Agent Runner] --> NativeHook[Native Hook or Plugin Event]
  NativeHook --> Adapter[agent-hub hook notify]
  Adapter --> Normalize[Normalize Runner Payload]
  Normalize --> Daemon[agent-hubd]
  Daemon --> Assign[Assign Envelope Fields]
  Assign --> Partition[Choose Partition from received_at]
  Partition --> Append[Append events/YYYY/MM/DD/events.jsonl]
  Append --> SessionState[Update Materialized Session]
  SessionState --> Ack[Return event_id, partition, offset]
```

## Consumer Flow

```mermaid
flowchart TD
  Consumer[Consumer] --> Fetch[agent-hub fetch --consumer-id --limit N]
  Fetch --> Daemon[agent-hubd]
  Daemon --> Cursor[Load Consumer Cursor]
  Cursor --> Read[Read Up To N Events]
  Read --> CrossPartition{Need Next Partition?}
  CrossPartition -- yes --> NextPartition[Open Next Date Partition]
  CrossPartition -- no --> Batch[Build Batch]
  NextPartition --> Batch
  Batch --> Advance{Peek?}
  Advance -- no --> SaveCursor[Save next_cursor]
  Advance -- yes --> KeepCursor[Do Not Save Cursor]
  SaveCursor --> Return[Return Events]
  KeepCursor --> Return
```

## Storage Boundary

```mermaid
flowchart LR
  subgraph Clients
    CLI[agent-hub CLI]
    Hooks[Hook Adapters]
    Consumers[Consumers]
  end

  subgraph DaemonProcess[agent-hubd]
    API[Local API]
    Validator[Validator]
    Queue[Queue Engine]
    SessionProjector[Session Projector]
  end

  subgraph Files[File Storage]
    Events[events/YYYY/MM/DD/events.jsonl]
    Index[events.idx]
    Cursors[consumers/*.cursor.json]
    Sessions[sessions/*/*.json]
  end

  CLI --> API
  Hooks --> API
  Consumers --> API
  API --> Validator
  Validator --> Queue
  Queue --> Events
  Queue --> Index
  Queue --> Cursors
  Queue --> SessionProjector
  SessionProjector --> Sessions
```

## Runner Integration Layers

```mermaid
flowchart TD
  subgraph CodexIntegration[Codex]
    CSessionStart[SessionStart] --> CMap[Codex Mapper]
    CPrompt[UserPromptSubmit] --> CMap
    CStop[Stop] --> CMap
    CPreTool[PreToolUse] --> CMap
    CPostTool[PostToolUse] --> CMap
  end

  subgraph OpencodeIntegration[opencode]
    OSession[session.created] --> OMap[opencode Mapper]
    OMessage[message.updated] --> OMap
    OIdle[session.idle] --> OMap
    OError[session.error] --> OMap
    OToolBefore[tool.execute.before] --> OMap
    OToolAfter[tool.execute.after] --> OMap
  end

  CMap --> Normalized[Normalized Agent Hub Events]
  OMap --> Normalized
  Normalized --> Daemon[agent-hubd]
```

## Normalized Event Types

- `agent.session.started`
- `agent.prompt.submitted`
- `agent.session.updated`
- `agent.session.finished`
- `agent.session.failed`
- `agent.tool.started`
- `agent.tool.finished`
- `agent.permission.requested`
- `agent.permission.replied`

## API Transport

Unix socket is the primary v1 transport:

```text
.agent-hub/agent-hub.sock
```

An optional localhost HTTP listener can be added for dashboards and external
tools:

```sh
agent-hub daemon start --http 127.0.0.1:0
```

The daemon API should be transport-neutral internally so the CLI can use the
same request and response models for Unix socket and HTTP.

## Test Harness

```mermaid
flowchart LR
  MockConfig[mock-config.json] --> FakeCodex[cmd/fake-codex]
  MockConfig --> FakeOpencode[cmd/fake-opencode]
  FakeCodex --> HookCommand[agent-hub hook notify]
  FakeOpencode --> HookCommand
  HookCommand --> Daemon[agent-hubd in temp AGENT_HUB_HOME]
  Daemon --> TestLog[Temp Date-Partitioned Event Log]
  TestConsumer[Go Test] --> Fetch[agent-hub fetch --limit N]
  Fetch --> Daemon
```

Fake runners must ignore host Codex and opencode config. They only fire hooks
declared in `--mock-config`.

