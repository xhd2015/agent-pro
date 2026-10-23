# DSH session inspection

`agent-pro dsh` reads DeepSeek Harness session artifacts locally. It does not require a running DSH server, model credentials, or the `dsh` executable. It never resumes a session, acquires its writer lock, or repairs or migrates stored data.

## Commands

```sh
agent-pro dsh sessions --here
agent-pro dsh sessions --dir /path/to/project --limit 10 --json
agent-pro dsh session list
agent-pro dsh session info <session-id> --json
agent-pro dsh session status <session-id>
agent-pro dsh session log <session-id> --tail 20
agent-pro dsh session log <session-id> --json
agent-pro dsh session messages <session-id>
agent-pro dsh session prompts <session-id>
agent-pro dsh session prompts --limit 5
```

Every command level accepts `-h` and `--help`. Results go to stdout; non-fatal `warning:` diagnostics and fatal `agent-pro:` errors go to stderr. Help and successful reads exit zero; invalid arguments and failed exact-session reads exit non-zero.

`session status` reports the last durable turn outcome, not process liveness. An incomplete turn can belong to a live process or a crashed one. A completed turn does not imply that its session owner has exited.

`session log` shows user and assistant messages, tool activity, and turn markers. `session messages` selects user and assistant messages. `session prompts` selects human-originated user messages; without an id it reads recent sessions. These are chronological views of recorded history, including inherited and superseded messages, not reconstructions of the current model context. Embedded assistant stream chunks are not printed a second time beside the settled assistant message. Message views render text blocks only; image, attachment, and other non-text content remains available in the raw NDJSON log.

`--json` on `session log` emits the selected artifact's complete durable event records as NDJSON, retaining event fields and plugin events. Other JSON commands return ordinary JSON documents. JSON output contains no terminal color escapes.

## Storage selection

The session root is selected in this order:

1. `--sessions-dir DIR`, used directly.
2. `--home DIR`, using `DIR/sessions`.
3. Non-blank `$DSH_HOME`, using `$DSH_HOME/sessions`.
4. `~/.dsh/sessions`.

Use `--sessions-dir` for DSH deployments whose persistence configuration points outside the standard home. The inspector does not execute Cordis configuration or infer custom roots from patch files. The `--here` and `--dir` filters compare cleaned absolute cwd strings; they do not resolve symlink aliases.

Each root contains project directories, then session directories. The reader selects the numerically highest canonical generation filename in each directory. It supports formats v3 and v4 in plain `session.vN.jsonl` and concatenated Zstandard `session.vN.jsonl.zstd` artifacts. Temporary and noncanonical filenames are not generations. An unsupported highest generation is never replaced by a readable older one. Roots mixing plain and compressed artifacts are rejected.

A missing root produces an empty listing. Unsupported or malformed session artifacts produce warnings during a listing; an exact inspection of such a session fails. A final plaintext fragment without a newline is omitted as an uncommitted tail. Damaged or truncated Zstandard streams fail rather than attempting DSH's compressed-tail recovery. Files remain unchanged on success and failure.

## Limits

- Only v3 and v4 are supported; v0, v1, and v2 need a DSH-owned migration before this inspector can interpret them. Newer generations are refused.
- The reader validates the records it uses, but does not implement DSH's full replay validation, plugin composition, historical migration, or current-context derivation.
- Inspection observes persisted bytes, not unflushed live events. It is not an atomic snapshot across multiple sessions.
- `send`, `open`, `focus`, `snapshot`, `fork`, and `backup` are not implemented. `ListLive` reports sessions whose `session.lock` write lease is open; it does not acquire that lock, and it does not require an iTerm tab.

## Development checks

From the agent-pro repository root:

```sh
go test ./agent/dsh/...
go -C cmd test ./agent-pro
```

The `cmd` directory is a separate Go module with a local replacement for the parent module. Tests use isolated fixture roots rather than the user's DSH home. The [decision note](notes/implemented/feature/2026-09-20-dsh-session-inspection.md) records the native-reader and no-live-control tradeoffs.
