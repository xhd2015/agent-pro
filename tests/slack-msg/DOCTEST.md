# slack-msg CLI (send | history | listen | channels | auth | session) Doctests

Doc-style tests for the unified first-class CLI at `cmd/slack-msg` with subcommands
`send`, `history`, `listen`, `channels` (`list` | `search`), `auth status`
(`--app` for app-level tokens), and `session` (`reply` | `history`). Ports contracts
from `tests/slack-send-cli` and `tests/slack-listen` (updated argv and `--token`
instead of `--bot-token`), history coverage (oldest→newest, `--json`, `--thread`),
channel list/search, auth status (bot via `auth.test`, app via `apps.connections.open`),
and session-bound list/info/update/reply/history (SeaTalk-like; channel top-level
posts, no `thread_ts`). Uses `pkgs/slackutil` for config/token/channel resolution and
`slacktest` + `SLACK_API_URL` for unit paths.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — user or doctest harness invoking the `slack-msg` binary.
- **slack-msg CLI** — root dispatcher for `send` | `history` | `listen` | `channels` |
  `auth` | `session`; top-level `-h`/`--help` lists the six commands plus Help
  topics (including `add-missing-scope`) and `slack-msg --help [--topic TOPIC]`;
  `--help --topic` / `--topic … --help` print a topic body; `--topic` alone or
  unknown topic → stderr + exit 1; unknown command → stderr + exit 1.
- **send** — posts via `chat.postMessage`; requires exactly one positional MESSAGE;
  flags `--token`, `--channel`, `--config`, `--thread`, `-h`/`--help`; token/channel
  from CLI → env → config; prints Sending / Using config / OK lines on success.
- **history** — fetches `conversations.history` or `conversations.replies` (`--thread`);
  prints human lines oldest→newest (`[ts] user: text`) or `--json` document in the
  same order; flags `--token`, `--channel`, `--config`, `--limit`, `--json`, `--thread`.
- **listen** — Socket Mode inbound bridge → filter → `pkgs/agentrunbridge`
  (`RunInteractiveOpen` for default thread session mode; `Run` + `CaptureStdout`
  for `--session-mode stateless`) → optional `chat.postMessage` (stateless agent
  body only; thread interactive open does not PostMessage agent stdout). Flags
  use `--token` (bot) and `--app-token`. **Default singleton lock** at
  `~/.agent-pro/slack-msg.listen.lock` when `--lock-file` omitted; `--no-lock`
  (or empty lock path) disables. After AuthTest, prints a **startup banner**
  (config path, team, bot identity, session-mode, require-mention, agent-runner,
  lock path or `(none)`). **Dedupe** inbound `app_mention`+`message` with the
  same `(channel, ts)` to one agent launch. **Operator logs** for accepted
  events (kind, user display, channel, ts, text) and agent open/errors. **Strip**
  bot `<@BOTID>` from agent prompt text. Thread mode **upserts** durable
  `sessions.json`, **appends** inbound lines to `messages.jsonl`, writes
  **SYSTEM.md** under `~/.agent-pro/slack-local-bot/sessions/<sessionID>/` with
  **session** CLI recipes (`slack-msg session history` / `session reply` — no
  raw `send --channel/--thread`), open-injects a prompt with session metadata +
  SYSTEM.md path, and launches `RunInteractiveOpen` with Env
  `SLACK_MSG_SESSION_ID` + `SLACK_MSG_CONFIG` (config only when path non-empty)
  as agent-run `-e KEY=VALUE`. **Stable session ids:** channel/group/MPIM →
  `slack-channel-{channelID}`; DM (`D…`) → `slack-dm-{userID}` (not per-message
  ts; event dedupe stays `channelID:ts`). Agent body is not PostMessaged in
  thread mode.
- **session** — session-bound management + agent recipes: `list` prints map rows
  (human header `SESSION_ID CHANNEL DIR UPDATED PREVIEW` with column padding,
  sort `updated_at` desc; empty map → empty stdout exit 0; `--json` / `--limit`);
  `info` shows one entry + derived `message_count` / `session_dir` (human
  `key: value` or `--json`); `update --dir PATH` sets workspace `dir` (must exist
  as directory; store absolute path; bump `updated_at`; preserve other fields;
  human `OK session=… dir=…` or `--json` full entry); `reply` posts
  `chat.postMessage` to the map entry's `channel_id` **without** `thread_ts`
  (channel top-level); `history` prints local `messages.jsonl` (oldest→newest;
  `--after-msg-id`, `--limit`, `--json`). Session id via `--session-id` or
  `SLACK_MSG_SESSION_ID`; config via `--config` or `SLACK_MSG_CONFIG` or map
  `config_path`. JSON list/info/update emit both `session_id` and
  `agent_session_id` (equal today). Empty `dir` human `-`, JSON `""`. Durable
  store: `~/.agent-pro/slack-local-bot/sessions.json` (optional entry field `dir`)
  + `sessions/<id>/messages.jsonl`.
- **channels** — `list` or `search QUERY` over `conversations.list` (paginated
  **per mapped type**); default `--types public,private` maps to
  `public_channel` + `private_channel`; default exclude archived; sort by name
  ascending; human lines `id  #name  kind  member|-` (two spaces between columns);
  `--json` shape `{"channels":[{id,name,is_private,is_member,is_archived}]}`;
  search default case-insensitive contains (strip leading `#`); `--exact` /
  `--prefix` mutually exclusive (both set → stderr + exit 1); no matches → empty
  human or `{"channels":[]}` with exit 0; flags `--token`, `--config`, `--types`,
  `--limit`, `--json`. **Scope soft-fail:** when multi-type and one type returns
  `missing_scope`, skip that type, emit short stderr warning
  (`warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope`
  when `needed` is present), continue with successful types, exit 0. **Scope
  hard-fail:** sole requested type `missing_scope`, or all types fail with no
  success → `channels failed: missing_scope (needed <needed>); see: slack-msg --help --topic add-missing-scope`
  (or without needed when absent, still with `see:` when scope-related), exit 1.
  Other API errors (e.g. `internal_error`, `invalid_auth`) hard fail with
  `channels failed:` (no soft-degrade; `see:` pointer not required).
- **auth** — `status` inspects bot token (default) or app token (`--app`);
  always prints `Using config from: <abs-path>` or `Using config from: (none)`;
  bot path calls `auth.test` and prints human lines `kind` / `ok` / `team` /
  `user` / `bot_id` / `url` / `token` (masked) or `--json` document; app path
  validates via `apps.connections.open` (Socket Mode / connections) and prints
  `kind: app`, `ok`, masked `token`, and a fixed `note` line; flags `--token`,
  `--app-token`, `--config`, `--json`, `--app`, `-h`/`--help`. Missing token →
  `bot token required` / `app token required`; API not ok → `auth failed:` prefix,
  exit 1. Stdout must never contain the full raw token.
- **SlackConfig JSON** — optional file with `botToken`, `appToken`, `defaultChannelId`,
  `knownChannels`; loaded only when `--config` or `SLACK_CONFIG` is set (no cwd walk).
- **Slack Web API** — real `slack.com` in integration; `slacktest` when `SLACK_API_URL`
  is set (`conversations.list`, `chat.postMessage`, `conversations.history`,
  `conversations.replies`, `auth.test`, `apps.connections.open`, Socket Mode).
- **agent-run** — external runner for listen; mocked via `SLACK_LISTEN_AGENT_RUN`.
- **Singleton lock** — default path `~/.agent-pro/slack-msg.listen.lock`; second
  listen instance exits non-zero with `another slack-msg is already running`
  when lock held; `--no-lock` disables singleton.
- **Local-bot session store** — under `~/.agent-pro/slack-local-bot/`:
  `sessions.json` (conversation map), `sessions/<sessionID>/SYSTEM.md` (playbook
  + `slack-msg session history` / `session reply` recipes; no secrets),
  `sessions/<sessionID>/messages.jsonl` (inbound/outbound log for session history).
- **Test harness** — builds `./cmd/slack-msg` once per session; quick-exit CLI runs
  for send/history/channels/auth/session/root; daemon probes for listen unit paths.
  Daemon leaves auto-isolate `--lock-file` under WorkDir unless `UseDefaultLock`
  or `NoLock` is set; optional `HomeDir` isolates `HOME` for default lock + session
  store; `CapturePosts` captures `chat.postMessage` on simple (non-daemon) runs
  for session reply leaves.

**Behaviors**

- `slack-msg -h|--help` — lists `send`, `history`, `listen`, `channels`, `auth`,
  `session`; Help topics including `add-missing-scope`; usage line
  `slack-msg --help [--topic TOPIC]`; exit 0; trailing `\n`.
- `slack-msg --help --topic add-missing-scope` / `--topic add-missing-scope --help`
  — topic guideline body (groups:read, reinstall, botToken/config); exit 0; trailing `\n`.
- `slack-msg --topic …` without `--help` — stderr e.g. `--topic requires --help`; exit 1.
- `slack-msg --help --topic <unknown>` — stderr `unknown help topic: …`; exit 1.
- `slack-msg <unknown>` — stderr error; exit 1.
- `slack-msg send [options] MESSAGE` — message required (reject 0 or 2+); missing
  token → `bot token required`; missing channel → `channel required`; success
  stdout ends with `OK ts=... channel=...` and trailing `\n`; API failure →
  `send failed:` prefix, exit 1, no `OK`.
- `slack-msg history [options] [CHANNEL]` — channel via flag/positional/env/config;
  human or JSON oldest→newest; API failure → `history failed:` prefix, exit 1.
- `slack-msg listen [options]` — missing bot/app token exit 1 before connect;
  config `(none)` vs absolute path logging; startup banner after AuthTest;
  filter bot-self / DM / mention / allowFrom / channel; dedupe same channel+ts;
  operator logs for accepted inbound + agent open/error; strip bot mention from
  prompt; thread vs stateless session routing via agentrunbridge; default lock
  path / `--no-lock`; thread mode sessions.json upsert + messages.jsonl append +
  SYSTEM.md (session recipes) + open inject + interactive open with
  `-e SLACK_MSG_SESSION_ID=…` / `-e SLACK_MSG_CONFIG=…` (no agent-body
  PostMessage); session id is `slack-channel-{channelID}` or
  `slack-dm-{userID}`; stateless replies PostMessage with `thread_ts` and optional
  `--reply-prefix` (no SYSTEM.md / session map).
- `slack-msg session list [--json] [--limit N]` — map rows sorted by `updated_at`
  desc; human table (header first label `SESSION_ID`); empty map empty stdout
  exit 0; `--json` `{"sessions":[…]}` with `session_id`+`agent_session_id`.
- `slack-msg session info [--session-id ID] [--json]` — one entry +
  `message_count` + `session_dir`; missing id / not found → exit 1.
- `slack-msg session update [--session-id ID] --dir PATH [--json]` — require
  existing session and `--dir` (exists, is directory); store abs `dir`; human
  `OK session=… dir=…` or `--json` full entry; errors for missing id / not found /
  nothing to update / dir missing or not a directory.
- `slack-msg session reply [options] MESSAGE` — resolve session id → map entry →
  config/token → `chat.postMessage` to `channel_id` **without** `thread_ts` →
  append outbound to messages.jsonl → `OK ts=… channel=…` + trailing `\n`;
  missing session id / unknown session / missing token → exit 1.
- `slack-msg session history [options]` — local messages.jsonl oldest→newest;
  `--after-msg-id` filters; `--json` optional; trailing `\n` on success.
- `slack-msg channels list [options]` — fetch per type, merge, exclude archived,
  sort by name, human or JSON; missing token → `bot token required`; multi-type
  private `missing_scope` → public rows + soft warning (with `see:` topic) exit 0;
  sole-type or all-type `missing_scope` → `channels failed:` + `see:` topic exit 1;
  other API fail → `channels failed:` exit 1.
- `slack-msg channels search [options] QUERY` — QUERY required; same list/scope
  rules then contains/exact/prefix filter; no matches exit 0 (still emit soft
  warning if a type was skipped); `--exact`+`--prefix` → error.
- `slack-msg auth status [options]` — default bot mode: resolve bot token
  (`--token` / `SLACK_BOT_TOKEN` / config `botToken`), call `auth.test`, print
  status human or `--json`; exit 0 when ok; exit 1 missing token or API fail.
- `slack-msg auth status --app [options]` — app mode: resolve app token
  (`--app-token` / `SLACK_APP_TOKEN` / config `appToken`), call
  `apps.connections.open`, print app status human or `--json`; exit 0 when ok.
- Token masking — human `token:` and JSON `token_masked` use type prefix +
  `...` + last 4 characters (e.g. `xoxb-...oken`); never emit full secret.

## Version

0.0.2

## Decision Tree

```
tests/slack-msg/
├── DOCTEST.md
├── SETUP.md                           # build binary, slacktest, mock agent, daemon helpers
├── testdata/
│   ├── valid-config.json
│   ├── empty-token-config.json
│   ├── empty-app-token-config.json
│   └── default-channel-name.json
├── help/                              # top-level -h / --help / --topic
│   ├── SETUP.md
│   ├── short-flag/
│   ├── long-flag/
│   └── topic/
│       ├── SETUP.md
│       ├── add-missing-scope/         # --help --topic / --topic --help
│       │   ├── SETUP.md
│       │   ├── help-then-topic/
│       │   └── topic-then-help/
│       ├── topic-alone/               # --topic without --help → exit 1
│       └── unknown-topic/             # --help --topic <unknown> → exit 1
├── unknown-command/
│   ├── SETUP.md
│   └── not-recognized/
├── send/
│   ├── SETUP.md
│   ├── help/
│   ├── message-errors/
│   ├── token-errors/
│   ├── channel-errors/
│   ├── config-errors/
│   ├── config-none/
│   ├── config-with-default/
│   ├── channel-resolve/               # label: unit
│   ├── send-success/                  # label: unit (incl. --thread)
│   ├── send-errors/                   # label: unit
│   └── integration/                   # label: integration, slow
├── history/
│   ├── SETUP.md
│   ├── help/
│   ├── token-errors/
│   ├── channel-errors/
│   ├── order-oldest-first/            # label: unit
│   ├── json-output/                   # label: unit
│   ├── limit/                         # label: unit
│   ├── thread-replies/                # label: unit
│   ├── channel-resolve/               # label: unit
│   └── history-errors/                # label: unit
├── listen/
│   ├── SETUP.md
│   ├── help/                          # includes --no-lock
│   ├── token-errors/                  # --token / --app-token
│   ├── lock/                          # label: unit (explicit / default / no-lock)
│   ├── banner/                        # label: unit (startup identity)
│   ├── config/
│   ├── filter/                        # label: unit
│   ├── dedupe/                        # label: unit (app_mention+message same ts)
│   ├── logs/                          # label: unit (operator inbound / agent error)
│   ├── session-routing/               # label: unit (+ SYSTEM.md / map / env inject)
│   ├── reply/                         # label: unit
│   └── integration/                   # label: integration, slow
├── session/                           # session list | info | update | reply | history
│   ├── SETUP.md
│   ├── help/                          # session -h / --help (list,info,update,reply,history)
│   ├── list/
│   │   ├── SETUP.md
│   │   └── success/                   # label: unit (empty / sorted human / json / limit)
│   ├── info/
│   │   ├── SETUP.md
│   │   ├── success/                   # label: unit (human keys / json / env session id)
│   │   └── errors/                    # missing/unknown session
│   ├── update/
│   │   ├── SETUP.md
│   │   ├── success/                   # label: unit (set-dir abs / preserves / json)
│   │   └── errors/                    # id / not found / missing dir / bad path
│   ├── reply/
│   │   ├── SETUP.md
│   │   ├── help/
│   │   ├── success/                   # label: unit (channel post no thread)
│   │   └── errors/                    # missing/unknown session, token
│   └── history/
│       ├── SETUP.md
│       ├── help/
│       ├── success/                   # label: unit (local messages.jsonl)
│       └── errors/
└── channels/
    ├── SETUP.md
    ├── help/                          # channels -h / --help
    ├── list/
    │   ├── SETUP.md
    │   ├── help/
    │   ├── human-sorted/              # label: unit (full scopes; excl. archived)
    │   ├── json-output/               # label: unit (full scopes)
    │   ├── limit/                     # label: unit
    │   ├── pagination/                # label: unit
    │   └── soft-scope/                # label: unit (private missing_scope soft-skip)
    ├── search/
    │   ├── SETUP.md
    │   ├── help/
    │   ├── contains/                  # label: unit (full scopes)
    │   ├── no-match/                  # label: unit (full scopes)
    │   ├── exact/                     # label: unit
    │   ├── prefix/                    # label: unit
    │   ├── soft-scope/                # label: unit (private missing_scope soft-skip)
    │   ├── missing-query/
    │   └── flags-conflict/
    ├── token-errors/
    ├── channels-errors/               # label: unit (hard fail: api / sole private / all types)
    └── unknown-subcommand/
└── auth/
    ├── SETUP.md
    ├── help/                          # auth -h / --help
    └── status/
        ├── SETUP.md
        ├── help/                      # auth status -h / --help
        ├── bot/                       # default mode (auth.test)
        │   ├── SETUP.md
        │   ├── success/               # label: unit (config / none / json)
        │   ├── token-errors/
        │   └── auth-errors/           # label: unit (auth.test fail)
        └── app/                       # --app (apps.connections.open)
            ├── SETUP.md
            ├── success/               # label: unit (token / json)
            └── token-errors/
```

Parameter ranking (most → least significant):

1. **Command / outcome** — top-level help vs unknown vs send vs history vs listen vs channels vs auth vs session
2. **Within command** — help vs validation-error vs success/unit path vs integration
3. **Session action** — `list` vs `info` vs `update` vs `reply` vs `history`
4. **Auth mode** — bot (default) vs `--app`
5. **Credential / session / config source** — CLI flags vs env (`SLACK_MSG_*`) vs map / `--config` JSON
6. **Backend** — slacktest (`SLACK_API_URL`) vs live Slack vs local session store vs no network

## Test Index

| # | Leaf | Labels | Description |
|---|------|--------|-------------|
| 1 | `help/short-flag` | (default) | Top-level `-h` lists commands including `auth`, `session` + Help topics (`add-missing-scope`); exit 0 |
| 2 | `help/long-flag` | (default) | Top-level `--help` same contract |
| 2a | `help/topic/add-missing-scope/help-then-topic` | (default) | `--help --topic add-missing-scope` guideline; exit 0 |
| 2b | `help/topic/add-missing-scope/topic-then-help` | (default) | `--topic add-missing-scope --help` same body; exit 0 |
| 2c | `help/topic/topic-alone` | (default) | `--topic` without `--help` → stderr requires --help; exit 1 |
| 2d | `help/topic/unknown-topic` | (default) | `--help --topic not-a-topic` → unknown help topic; exit 1 |
| 3 | `unknown-command/not-recognized` | (default) | Unknown subcommand → stderr; exit 1 |
| 4 | `send/help/short-flag` | (default) | `send -h` usage; exit 0 |
| 5 | `send/help/long-flag` | (default) | `send --help` same as `-h` |
| 6 | `send/message-errors/missing-message` | (default) | No MESSAGE → `message required` |
| 7 | `send/message-errors/multiple-positionals` | (default) | Two positionals → `exactly one message required` |
| 8 | `send/token-errors/missing-token` | (default) | No bot token → `bot token required` |
| 9 | `send/channel-errors/missing-channel` | (default) | No channel → `channel required` |
| 10 | `send/config-errors/bad-config-path` | (default) | Missing config file → `failed to load config` |
| 11 | `send/config-errors/empty-bot-token` | (default) | Empty config botToken → `botToken is empty in` |
| 12 | `send/config-none/stdout-line` | (default) | CLI flags only → `Using config from: (none)` |
| 13 | `send/channel-resolve/api-name-with-hash` | unit | `#general` via conversations.list |
| 14 | `send/channel-resolve/api-name-without-hash` | unit | `general` normalized via API |
| 15 | `send/channel-resolve/direct-channel-id` | unit | `C...` passthrough |
| 16 | `send/channel-resolve/direct-dm-id` | unit | `D...` passthrough |
| 17 | `send/channel-resolve/direct-group-id` | unit | `G...` passthrough |
| 18 | `send/channel-resolve/config-known-channels` | unit | knownChannels fast path |
| 19 | `send/send-success/cli-flags` | unit | `--token --channel` + message |
| 20 | `send/send-success/channel-by-id` | unit | Channel ID + custom message |
| 21 | `send/send-success/multi-word-message` | unit | Single multi-word MESSAGE |
| 22 | `send/send-success/env-token` | unit | `SLACK_BOT_TOKEN` fallback |
| 23 | `send/send-success/env-channel` | unit | `SLACK_CHANNEL` fallback |
| 24 | `send/send-success/thread-ts` | unit | `--thread TS` post in thread |
| 25 | `send/send-errors/channel-not-found` | unit | Unknown name → `channel not found` |
| 26 | `send/send-errors/api-post-failed` | unit | PostMessage error → `send failed:` |
| 27 | `send/config-with-default/message-only` | unit | `--config` + message uses defaults |
| 28 | `send/config-with-default/override-channel` | unit | CLI `--channel` wins over config |
| 29 | `send/config-with-default/default-channel-name` | unit | Name-shaped defaultChannelId |
| 30 | `send/integration/live-explicit-config` | integration, slow | Live send via repo config |
| 31 | `history/help/short-flag` | (default) | `history -h`; exit 0 |
| 32 | `history/help/long-flag` | (default) | `history --help` |
| 33 | `history/token-errors/missing-token` | (default) | No bot token → `bot token required` |
| 34 | `history/channel-errors/missing-channel` | (default) | No channel → `channel required` |
| 35 | `history/order-oldest-first/multi-message` | unit | Human lines oldest→newest |
| 36 | `history/json-output/chronological` | unit | JSON messages oldest→newest |
| 37 | `history/limit/limits-results` | unit | `--limit 2` trims to 2 chronological lines |
| 38 | `history/thread-replies/prints-replies` | unit | `--thread` uses replies path |
| 39 | `history/channel-resolve/direct-channel-id` | unit | Direct channel ID history |
| 40 | `history/channel-resolve/api-name-with-hash` | unit | `#general` resolve then history |
| 41 | `history/history-errors/channel-not-found` | unit | Unknown channel → `history failed:` + not found |
| 42 | `history/history-errors/api-failed` | unit | API error → `history failed:` |
| 43 | `listen/help/short-flag` | (default) | `listen -h`; `--token` not `--bot-token`; lists `--no-lock` |
| 44 | `listen/help/long-flag` | (default) | `listen --help` same options including `--no-lock` |
| 45 | `listen/token-errors/missing-bot-token` | (default) | No bot token → `bot token required` |
| 46 | `listen/token-errors/missing-app-token` | (default) | No app token → `app token required` |
| 47 | `listen/lock/already-running` | unit | Explicit `--lock-file`: second instance → `another slack-msg is already running` |
| 47a | `listen/lock/default-path` | unit | No lock flag → default `~/.agent-pro/slack-msg.listen.lock`; second conflicts |
| 47b | `listen/lock/no-lock` | unit | `--no-lock` → banner lock `(none)`; second has no singleton conflict |
| 47c | `listen/banner/startup-identity` | unit | Startup banner: team, bot id, session-mode, require-mention, agent-runner, lock |
| 48 | `listen/config/none-stdout-line` | unit | `Using config from: (none)` |
| 49 | `listen/config/explicit-path` | unit | Absolute `--config` path logged |
| 50 | `listen/config/bad-config-path` | (default) | Missing config → `failed to load config` |
| 51 | `listen/filter/ignore-bot-self` | unit | Bot-authored message ignored |
| 52 | `listen/filter/dm-always-processed` | unit | DM without mention processed |
| 53 | `listen/filter/channel-requires-mention` | unit | Channel msg without mention ignored |
| 54 | `listen/filter/channel-no-require-mention` | unit | `--no-require-mention` processes channel |
| 55 | `listen/filter/allow-from-blocked` | unit | User not in allowFrom ignored |
| 56 | `listen/filter/allow-from-wildcard` | unit | Default allow processes any user |
| 57 | `listen/filter/channel-filter-excludes` | unit | `--channel` filter drops others |
| 57a | `listen/dedupe/same-channel-ts` | unit | `app_mention`+`message` same channel+ts → 1 agent launch |
| 57b | `listen/dedupe/different-ts` | unit | Two different ts → 2 agent launches |
| 57c | `listen/logs/accepted-inbound` | unit | Accepted event logs kind/user display/channel/ts/text + agent open |
| 57d | `listen/logs/agent-open-error` | unit | Agent open failure is logged (not silent) |
| 58 | `listen/session-routing/thread-first-run` | unit | First msg → RunInteractiveOpen (`run` + `--session-id=slack-channel-{C}` + auto-send/new-terminal/open) |
| 59 | `listen/session-routing/thread-follow-up-open` | unit | Follow-up also RunInteractiveOpen `run` (not `send`; was `thread-follow-up-send`) |
| 59a | `listen/session-routing/channel-stable-session` | unit | Two channel msgs different ts → same `slack-channel-{C}` session id |
| 59b | `listen/session-routing/dm-session-key` | unit | DM → `--session-id=slack-dm-{userID}` (not `slack-channel-D…`) |
| 60 | `listen/session-routing/stateless-each-run` | unit | Stateless → every msg `Run` capture (no `--open` / session-id) |
| 60a | `listen/session-routing/thread-system-md` | unit | Thread first open writes SYSTEM.md under `slack-channel-{C}` with **session** recipes (no raw send --channel/--thread) |
| 60b | `listen/session-routing/thread-open-inject` | unit | Open prompt: inject markers, stripped mention, from:, SYSTEM.md path |
| 60c | `listen/session-routing/stateless-no-system-md` | unit | Stateless does not write SYSTEM.md |
| 60d | `listen/session-routing/thread-session-map` | unit | Upserts `sessions.json` entry (`slack-channel-{C}`, channel, thread_ts, config_path) |
| 60e | `listen/session-routing/thread-messages-log` | unit | Appends inbound line to `messages.jsonl` under stable session id |
| 60f | `listen/session-routing/thread-env-flags` | unit | Agent argv includes `-e SLACK_MSG_SESSION_ID=slack-channel-{C}` and `-e SLACK_MSG_CONFIG=…` |
| 61 | `listen/reply/posts-in-thread` | unit | Thread interactive open → no agent-body PostMessage |
| 62 | `listen/reply/reply-prefix` | unit | Stateless `--reply-prefix` + agent body PostMessage with `thread_ts` |
| 63 | `listen/integration/live-socket-reply` | integration, slow | Live Socket Mode connect probe |
| 63a | `session/help/short-flag` | (default) | `session -h` lists list/info/update/reply/history; exit 0 |
| 63b | `session/help/long-flag` | (default) | `session --help` same as `-h` |
| 63r | `session/list/success/empty-map` | unit | Empty sessions.json → empty stdout; exit 0 |
| 63s | `session/list/success/multi-sorted-human` | unit | Human header `SESSION_ID…` + column padding; sort updated_at desc; empty dir `-` |
| 63t | `session/list/success/json-output` | unit | `--json` sessions[] with session_id + agent_session_id; sort desc |
| 63u | `session/list/success/limit` | unit | `--limit N` after sort; at most N data rows |
| 63v | `session/info/success/human-keys` | unit | Human key: value incl. agent_session_id, message_count, session_dir |
| 63w | `session/info/success/json-output` | unit | `--json` object with session_id + agent_session_id + message_count + session_dir |
| 63x | `session/info/success/env-session-id` | unit | `SLACK_MSG_SESSION_ID` without `--session-id` |
| 63y | `session/info/errors/missing-session-id` | (default) | No session id → session id required; exit 1 |
| 63z | `session/info/errors/unknown-session` | (default) | Unknown id → session not found; exit 1 |
| 64a | `session/update/success/set-dir` | unit | `--dir` existing dir → store abs path; OK line; bump updated_at; preserve fields |
| 64b | `session/update/success/json-output` | unit | `update --json` full map entry with session_id + agent_session_id + dir |
| 64c | `session/update/errors/missing-session-id` | (default) | No session id → exit 1 |
| 64d | `session/update/errors/unknown-session` | (default) | Unknown session → exit 1 |
| 64e | `session/update/errors/missing-dir` | (default) | No `--dir` → nothing to update; exit 1 |
| 64f | `session/update/errors/dir-does-not-exist` | (default) | `--dir` missing path → dir does not exist; exit 1 |
| 64g | `session/update/errors/dir-not-directory` | (default) | `--dir` is a file → dir is not a directory; exit 1 |
| 63c | `session/reply/help/short-flag` | (default) | `session reply -h`; lists --session-id / --config / SLACK_MSG_* |
| 63d | `session/reply/help/long-flag` | (default) | `session reply --help` |
| 63e | `session/reply/success/map-config-flag` | unit | Map+`--session-id`+`--config` → PostMessage channel only (no thread_ts); OK line |
| 63f | `session/reply/success/env-session-and-config` | unit | `SLACK_MSG_SESSION_ID` + `SLACK_MSG_CONFIG` resolve; channel post no thread |
| 63g | `session/reply/success/appends-outbound-log` | unit | Successful reply appends `direction=out` to messages.jsonl |
| 63h | `session/reply/errors/missing-session-id` | (default) | No session id → stderr session id required; exit 1 |
| 63i | `session/reply/errors/unknown-session` | (default) | Unknown session → session not found; exit 1 |
| 63j | `session/reply/errors/missing-token` | (default) | Map without config/token → bot token / config error; exit 1 |
| 63k | `session/history/help/short-flag` | (default) | `session history -h` |
| 63l | `session/history/help/long-flag` | (default) | `session history --help` |
| 63m | `session/history/success/chronological` | unit | Local log lines oldest→newest |
| 63n | `session/history/success/after-msg-id` | unit | `--after-msg-id` prints only later messages |
| 63o | `session/history/success/json-output` | unit | `--json` document; trailing `\n` |
| 63p | `session/history/errors/missing-session-id` | (default) | No session id → exit 1 |
| 63q | `session/history/errors/unknown-session` | (default) | Unknown session → exit 1 |
| 64 | `channels/help/short-flag` | (default) | `channels -h`; lists list/search |
| 65 | `channels/help/long-flag` | (default) | `channels --help` |
| 66 | `channels/list/help/short-flag` | (default) | `channels list -h` |
| 67 | `channels/list/help/long-flag` | (default) | `channels list --help` |
| 68 | `channels/search/help/short-flag` | (default) | `channels search -h` |
| 69 | `channels/search/help/long-flag` | (default) | `channels search --help` |
| 70 | `channels/list/human-sorted/multi-channel` | unit | Human lines name-sorted; member column; excl. archived |
| 71 | `channels/list/json-output/sorted` | unit | JSON channels array sorted by name; excl. archived |
| 72 | `channels/list/limit/limits-results` | unit | `--limit 2` prints first two by name |
| 73 | `channels/list/pagination/merges-pages` | unit | Cursor pages merged then sorted |
| 74 | `channels/search/contains/case-insensitive` | unit | QUERY contains match (case-insensitive; strip `#`) |
| 75 | `channels/search/no-match/empty-exit-0` | unit | No hits → empty / `{"channels":[]}`; exit 0 |
| 76 | `channels/search/exact/name-only` | unit | `--exact` exact name only |
| 77 | `channels/search/prefix/name-prefix` | unit | `--prefix` prefix only |
| 78 | `channels/search/missing-query/requires-query` | (default) | No QUERY → stderr; exit 1 |
| 79 | `channels/search/flags-conflict/exact-and-prefix` | (default) | Both `--exact` and `--prefix` → exit 1 |
| 80 | `channels/token-errors/missing-token` | (default) | No bot token → `bot token required` |
| 81 | `channels/channels-errors/api-failed` | unit | conversations.list error → `channels failed:` |
| 82 | `channels/channels-errors/private-only-missing-scope` | unit | `--types private` + missing_scope → hard fail with needed= + `see:` topic |
| 83 | `channels/channels-errors/all-types-missing-scope` | unit | All types missing_scope → hard fail exit 1 + `see:` topic |
| 84 | `channels/list/soft-scope/private-missing` | unit | Default multi-type; private missing_scope → public human + warning + `see:` |
| 85 | `channels/list/soft-scope/private-missing-json` | unit | Same soft-skip with `--json` + `see:` |
| 86 | `channels/search/soft-scope/public-hit` | unit | Search hits public after soft-skip private; warning + `see:` |
| 87 | `channels/search/soft-scope/no-match` | unit | Search empty after soft-skip; warning + `see:`; exit 0 |
| 88 | `channels/unknown-subcommand/not-recognized` | (default) | Unknown channels subcommand → exit 1 |
| 89 | `auth/help/short-flag` | (default) | `auth -h` lists `status`; exit 0 |
| 90 | `auth/help/long-flag` | (default) | `auth --help` same as `-h` |
| 91 | `auth/status/help/short-flag` | (default) | `auth status -h` lists `--app` and flags; exit 0 |
| 92 | `auth/status/help/long-flag` | (default) | `auth status --help` same as `-h` |
| 93 | `auth/status/bot/success/with-config` | unit | Bot status + `--config` abs path; auth.test fields; masked token |
| 94 | `auth/status/bot/success/no-config` | unit | Bot status `(none)` + `--token`; masked token; exit 0 |
| 95 | `auth/status/bot/success/json` | unit | Bot `--json` document; no raw token; trailing `\n` |
| 96 | `auth/status/bot/token-errors/missing-token` | (default) | No bot token → `bot token required`; exit 1 |
| 97 | `auth/status/bot/auth-errors/api-failed` | unit | auth.test fail → `auth failed:`; exit 1 |
| 98 | `auth/status/app/success/with-token` | unit | `--app` + `--app-token`; kind app; note line; exit 0 |
| 99 | `auth/status/app/success/json` | unit | `--app --json` document; masked token; exit 0 |
| 100 | `auth/status/app/token-errors/missing-token` | (default) | `--app` no app token → `app token required`; exit 1 |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./tests/slack-msg                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./tests/slack-msg
doctest test --label-all ./tests/slack-msg

# Structure validation
doctest vet ./tests/slack-msg

# Default CI — help + validation (no labels; no network)
doctest test ./tests/slack-msg

# Unit suite — slacktest + mock agent-run
doctest test --label unit ./tests/slack-msg

# Live Slack
doctest test --label integration ./tests/slack-msg
doctest test --label slow ./tests/slack-msg

# Single leaf
doctest test -v ./tests/slack-msg/help/short-flag
doctest test -v ./tests/slack-msg/help/topic/add-missing-scope/help-then-topic
doctest test -v ./tests/slack-msg/history/order-oldest-first/multi-message
doctest test -v ./tests/slack-msg/channels/list/human-sorted/multi-channel
doctest test -v ./tests/slack-msg/channels/list/soft-scope/private-missing
doctest test -v ./tests/slack-msg/auth/status/bot/success/no-config
doctest test -v ./tests/slack-msg/auth/status/app/success/with-token
doctest test -v ./tests/slack-msg/listen/lock/default-path
doctest test -v ./tests/slack-msg/listen/banner/startup-identity
doctest test -v ./tests/slack-msg/listen/dedupe/same-channel-ts
doctest test -v ./tests/slack-msg/listen/session-routing/thread-open-inject
doctest test -v ./tests/slack-msg/listen/session-routing/thread-session-map
doctest test -v ./tests/slack-msg/listen/session-routing/thread-env-flags
doctest test -v ./tests/slack-msg/listen/session-routing/channel-stable-session
doctest test -v ./tests/slack-msg/listen/session-routing/dm-session-key
doctest test -v ./tests/slack-msg/slacksession/reply/success/map-config-flag
doctest test -v ./tests/slack-msg/slacksession/history/success/chronological
doctest test -v ./tests/slack-msg/slacksession/list/success/multi-sorted-human
doctest test -v ./tests/slack-msg/slacksession/info/success/human-keys
doctest test -v ./tests/slack-msg/slacksession/update/success/set-dir
```

**Implementer note (`session list` / `info` / `update` — RED until landed):**

1. Help lists commands: `list`, `info`, `update`, `reply`, `history`.
2. Map entry optional `dir` (workspace); preserve across listen/reply upserts when omitted.
3. `list`: sort `updated_at` desc; human columns `SESSION_ID CHANNEL DIR UPDATED PREVIEW`
   with right-pad to max width, join columns with two spaces; empty dir → `-`;
   empty map → empty stdout exit 0; `--limit` after sort; `--json` array entries emit
   both `session_id` and `agent_session_id` (same value today).
4. `info`: human `key: value` lines; derived `message_count` (jsonl lines),
   `session_dir` (`…/sessions/<id>`); JSON same keys; empty dir human `-` / JSON `""`.
5. `update --dir PATH`: path must exist and be a directory; store absolute path;
   bump `updated_at`; preserve other fields; human `OK session=<id> dir=<abs>\n`;
   `--json` full updated entry including both session id fields.
6. Errors (stderr, empty stdout, exit 1): `session id required`, `session not found`,
   `nothing to update`, `dir does not exist`, `dir is not a directory`.

**Implementer note (stable session keys — RED until `sessionID` updated):**

1. Channel / group / MPIM: `session_id = "slack-channel-" + channelID`.
2. DM (`isDirectMessage`): `session_id = "slack-dm-" + userID` (not `slack-channel-D…`).
3. Keep event dedupe as `channelID:ts`. Keep `thread_ts` on map as metadata only.
4. No migration of old `slack-{C}-{ts}` dirs (orphan OK).
5. Group/MPIM use channel key (not DM).

**Implementer note (`session reply` / `session history` / listen bridge):**

1. Root help lists `session` with send/history/listen/channels/auth.
2. Durable store under `$HOME/.agent-pro/slack-local-bot/`: `sessions.json` map +
   `sessions/<id>/{SYSTEM.md,messages.jsonl}` (ids use stable keys above).
3. `session reply`: resolve session id (`--session-id` / `SLACK_MSG_SESSION_ID`) →
   map entry → config (`--config` / `SLACK_MSG_CONFIG` / map.config_path) → bot token →
   `chat.postMessage` **without** thread_ts → append outbound log → `OK ts=… channel=…\n`.
4. `session history`: read local messages.jsonl; `--after-msg-id` / `--limit` / `--json`.
5. Listen thread: upsert map (abs config_path), append inbound log, SYSTEM.md recipes
   use `slack-msg session history` / `session reply` only (no `send --channel/--thread`),
   `RunInteractiveOpen` Env: `SLACK_MSG_SESSION_ID` + `SLACK_MSG_CONFIG` → agent-run `-e`.
6. Harness: `HomeDir` for CLI + daemon; `CapturePosts` for session reply PostMessage asserts;
   `insertConfigAfterSubcommand` places `--config` after `session reply|history`.

**Implementer note (`auth status` RED until done):**

1. Root help lists `auth      Inspect bot or app token status` with the other
   commands (update golden help ASSERT carefully).
2. `slack-msg auth status` (bot): resolve bot token CLI → env → config; always
   print `Using config from:` line first; call `auth.test` via `SLACK_API_URL`
   hook; human fields locked in leaves; mask token as type prefix + `...` +
   last 4 chars; trailing `\n`; exit 0 on ok.
3. `slack-msg auth status --app`: resolve app token CLI → env → config; validate
   with **`apps.connections.open`** (Socket Mode / connections; mockable via
   default slacktest); human lines include fixed `note:`; exit 0 on ok.
4. Errors: missing bot → `bot token required`; missing app → `app token required`;
   API fail → `auth failed:` prefix; exit 1. Never print full raw token on stdout.
5. `--json`: `json.Encoder` document with `config`, `kind`, `ok`, identity
   fields (bot), `token_masked`, and app `note` when kind is app; trailing `\n`.
6. Harness: default slacktest overrides `/auth.test` with bot_id fixture;
   `AuthAPIFail` server returns auth.test `invalid_auth`. `insertConfigAfterSubcommand`
   places `--config` after `auth status`.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/slacktest"

	"github.com/xhd2015/doctest/session"
)

const (
	slackTestToken         = "xoxb-slacktest-token"
	slackTestBotToken      = slackTestToken
	slackTestAppToken      = "xapp-slacktest-token"
	slackTestChannelID     = "C0ALE44K5J6"
	slackTestDMChannelID   = "D024BE91L"
	slackTestUserID        = "W012A3CDE"
	slackTestOtherUserID   = "W0OTHERUSR"
	slackTestBotUserID     = "U023BECGF"
	slackTestBotName       = "TestSlackBot" // slacktest bots.info / BotName default
	slackTestTeamID        = "T024BE7LD"
	// auth.test fixture fields (override slacktest default; includes bot_id).
	slackTestTeamName      = "SlackTest Team"
	slackTestUserName      = "Egon Spengler"
	// users.info display_name for defaultNonBotUser (slackTestUserID).
	slackTestUserDisplayName = "spengler"
	slackTestAuthURL       = "https://localhost.localdomain/"
	slackTestAuthBotID     = "B0TESTBOTID"
	// Product defaults under $HOME (isolate via Request.HomeDir in unit leaves).
	defaultListenLockRelPath   = ".agent-pro/slack-msg.listen.lock"
	defaultSlackLocalBotRelDir = ".agent-pro/slack-local-bot"
	// Masked forms: type prefix + "..." + last 4 chars of full token.
	slackTestTokenMasked   = "xoxb-...oken"
	slackTestAppTokenMasked = "xapp-...oken"
	// valid-config.json botToken last-4 mask.
	validConfigBotTokenMasked = "xoxb-...oken"
	validConfigAppTokenMasked = "xapp-...oken"
	defaultAgentReply      = "mock agent reply"
	envAgentRun            = "SLACK_LISTEN_AGENT_RUN"
	envAgentLog            = "SLACK_LISTEN_AGENT_LOG"
)

// History fixture: API returns newest-first; CLI must print oldest→newest.
var historyMessagesNewestFirst = []map[string]any{
	{"type": "message", "user": "U_NEWER", "text": "second message", "ts": "1710000002.000200"},
	{"type": "message", "user": "U_OLDER", "text": "first message", "ts": "1710000001.000100"},
	{"type": "message", "user": "U_NEWEST", "text": "third message", "ts": "1710000003.000300"},
}

// Thread replies fixture (newest-first in mock; CLI chronological).
var threadRepliesNewestFirst = []map[string]any{
	{"type": "message", "user": "U_R2", "text": "reply two", "ts": "1710001000.000300", "thread_ts": "1710001000.000100"},
	{"type": "message", "user": "U_R1", "text": "reply one", "ts": "1710001000.000200", "thread_ts": "1710001000.000100"},
	{"type": "message", "user": "U_PARENT", "text": "parent", "ts": "1710001000.000100", "thread_ts": "1710001000.000100"},
}

type InjectedEvent struct {
	Kind     string // message | app_mention | dm
	Channel  string
	User     string
	Text     string
	TS       string
	ThreadTS string
}

type CapturedPost struct {
	Channel  string
	Text     string
	ThreadTS string
}

type Request struct {
	RepoRoot      string
	WorkDir       string
	Bin           string
	Args          []string
	ConfigPath    string
	ConfigFixture string
	ConfigInline  string
	SlackAPIURL   string
	Env           []string
	ClearSlackEnv bool

	// Listen-oriented fields (also used when ListenMode).
	ListenMode     bool
	BotToken       string
	AppToken       string
	LockFile       string
	NoLock         bool // pass --no-lock (disable singleton)
	UseDefaultLock bool // omit --lock-file/--no-lock; product default under HomeDir/HOME
	HomeDir        string // if set, env HOME=<HomeDir> for lock + session store isolation (CLI + daemon)
	Daemon         bool
	InjectEvents   []InjectedEvent
	WantAgentCalls int // -1 => len(InjectEvents); counts launch lines only (not tty status)
	WantPosts      int // >0 => wait for at least this many chat.postMessage captures
	CapturePosts   bool // simple CLI (session reply): capture chat.postMessage via dedicated slacktest
	AgentLogPath   string
	MockAgentPath  string
	ObserveTimeout time.Duration
	Posts          *[]CapturedPost
	SecondInstance bool

	// History server variants: default, historyFail.
	HistoryAPIFail bool
	// Auth server variants: default auth.test success (with bot_id), authFail.
	AuthAPIFail bool
	// Channels server variants: default, channelsFail, channelsPaginated,
	// channelsPrivateMissingScope (fail private_channel; public ok),
	// channelsAllMissingScope (missing_scope for every types request).
	ChannelsAPIFail              bool
	ChannelsPaginated            bool
	ChannelsPrivateMissingScope  bool
	ChannelsAllMissingScope      bool
}

type Response struct {
	ExitCode         int
	Stdout           string
	Stderr           string
	AgentInvocations []string
	PostMessages     []CapturedPost
	SecondExitCode   int
	SecondStderr     string
	SecondStdout     string
}

// Default conversations.list fixture (intentionally unsorted by name).
// Includes archived old-stuff (excluded by channels list/search by default).
// general/random IDs kept for send/history channel-resolve leaves.
var slackTestChannels = []slack.Channel{
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID:        "C0OTHERCHAN",
				IsPrivate: false,
			},
			Name:       "random",
			IsArchived: false,
		},
		IsMember: false,
	},
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID:        "C0ALE44K5J6",
				IsPrivate: false,
			},
			Name:       "general",
			IsArchived: false,
		},
		IsMember: true,
	},
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID:        "C0AGENTDBG1",
				IsPrivate: true,
			},
			Name:       "agent-pro-debug",
			IsArchived: false,
		},
		IsMember: false,
	},
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID:        "C0ARCHIVED1",
				IsPrivate: false,
			},
			Name:       "old-stuff",
			IsArchived: true,
		},
		IsMember: true,
	},
}

// Paginated list pages (merged server-side via cursor); CLI merges then sorts/filters.
var slackTestChannelsPage1 = []slack.Channel{
	slackTestChannels[0], // random
	slackTestChannels[2], // agent-pro-debug
}
var slackTestChannelsPage2 = []slack.Channel{
	slackTestChannels[1], // general
	slackTestChannels[3], // old-stuff (archived)
}

// Public-only subset for soft-scope fixtures (private_channel soft-skipped).
// Unsorted fixture order preserved from slackTestChannels indices 0,1,3.
var slackTestPublicChannels = []slack.Channel{
	slackTestChannels[0], // random
	slackTestChannels[1], // general
	slackTestChannels[3], // old-stuff (archived; still returned by API, CLI excludes)
}

// Expected non-archived sorted human lines (name asc; excl. archived):
//   C0AGENTDBG1  #agent-pro-debug  private  -
//   C0ALE44K5J6  #general  public  member
//   C0OTHERCHAN  #random  public  -
//
// Soft-scope (private missing) non-archived sorted public-only:
//   C0ALE44K5J6  #general  public  member
//   C0OTHERCHAN  #random  public  -

func findModuleRoot(start string) (string, error) {
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("go.mod not found above %s", start)
		}
	}
}

var (
	buildOnce sync.Once
	builtBin  string
	buildErr  error
)

func sessionCacheDir(sessionID string) string {
	return filepath.Join(os.TempDir(), "slack-msg-doctest-"+sessionID)
}

func withFileLock(lockPath string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func buildSlackMsg(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	buildOnce.Do(func() {
		repoRoot, err := findModuleRoot(d.DOCTEST_ROOT)
		if err != nil {
			buildErr = err
			return
		}
		cacheDir := sessionCacheDir(d.DOCTEST_SESSION_ID)
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			buildErr = err
			return
		}
		bin := filepath.Join(cacheDir, "slack-msg")
		ready := filepath.Join(cacheDir, "bin.ready")
		lock := filepath.Join(cacheDir, "build.lock")
		buildErr = withFileLock(lock, func() error {
			if fileExists(ready) && fileExists(bin) {
				builtBin = bin
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			// Build to a temp path then rename so parallel leaves never exec a
			// partially written binary (Linux ETXTBSY under repo-l2 load).
			tmpBin := bin + ".building"
			_ = os.Remove(tmpBin)
			cmd := exec.CommandContext(ctx, runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", tmpBin, "./slack-msg")
			cmd.Dir = repoRoot
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				_ = os.Remove(tmpBin)
				return fmt.Errorf("go build -C cmd ./slack-msg: %w\n%s", err, stderr.String())
			}
			if err := os.Rename(tmpBin, bin); err != nil {
				_ = os.Remove(tmpBin)
				return fmt.Errorf("install slack-msg binary: %w", err)
			}
			if err := os.WriteFile(ready, []byte("ok"), 0o644); err != nil {
				return err
			}
			builtBin = bin
			return nil
		})
	})
	return builtBin, buildErr
}

func resolveFixturePath(root, name string) string {
	candidates := []string{
		filepath.Join(root, "testdata", name),
		filepath.Join(root, name),
	}
	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}
	return filepath.Join(root, "testdata", name)
}

func materializeConfig(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req.ConfigInline == "" && req.ConfigFixture == "" {
		return nil
	}
	path := req.ConfigPath
	if path == "" {
		if req.WorkDir == "" {
			req.WorkDir = t.TempDir()
		}
		path = filepath.Join(req.WorkDir, "slack-config.json")
	}
	var data []byte
	var err error
	if req.ConfigInline != "" {
		data = []byte(req.ConfigInline)
	} else {
		data, err = os.ReadFile(resolveFixturePath(d.DOCTEST_ROOT, req.ConfigFixture))
		if err != nil {
			return fmt.Errorf("read fixture %s: %w", req.ConfigFixture, err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	req.ConfigPath = abs
	return nil
}

func mergeEnv(base []string, extra ...string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(base)+len(extra))
	for _, e := range base {
		if i := strings.IndexByte(e, '='); i > 0 {
			seen[e[:i]] = struct{}{}
		}
		out = append(out, e)
	}
	for _, e := range extra {
		if i := strings.IndexByte(e, '='); i > 0 {
			k := e[:i]
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}

func withoutEnvKeys(env []string, keys ...string) []string {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	out := make([]string, 0, len(env))
	for _, e := range env {
		if i := strings.IndexByte(e, '='); i > 0 {
			if _, drop := keySet[e[:i]]; drop {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

type slackTestMode int

const (
	slackTestDefault slackTestMode = iota
	slackTestPostFail
	slackTestHistoryFail
	slackTestAuthFail
	slackTestChannelsFail
	slackTestChannelsPaginated
	slackTestChannelsPrivateMissingScope
	slackTestChannelsAllMissingScope
)

var (
	// Per-process slacktest servers only. Do not cache API URLs on disk across
	// go-test packages — that reuses dead servers after the creating process exits.
	// slacktest.NewTestServer registers into an internal hub so the listener stays live.
	slackTestOnce                            sync.Once
	slackTestURL                             string
	slackTestErr                             error
	slackTestPostFailOnce                    sync.Once
	slackTestPostFailURL                     string
	slackTestPostFailErr                     error
	slackTestHistoryFailOnce                 sync.Once
	slackTestHistoryFailURL                  string
	slackTestHistoryFailErr                  error
	slackTestAuthFailOnce                    sync.Once
	slackTestAuthFailURL                     string
	slackTestAuthFailErr                     error
	slackTestChannelsFailOnce                sync.Once
	slackTestChannelsFailURL                 string
	slackTestChannelsFailErr                 error
	slackTestChannelsPaginatedOnce           sync.Once
	slackTestChannelsPaginatedURL            string
	slackTestChannelsPaginatedErr            error
	slackTestChannelsPrivateMissingScopeOnce sync.Once
	slackTestChannelsPrivateMissingScopeURL  string
	slackTestChannelsPrivateMissingScopeErr  error
	slackTestChannelsAllMissingScopeOnce     sync.Once
	slackTestChannelsAllMissingScopeURL      string
	slackTestChannelsAllMissingScopeErr      error
)

func conversationsListHandler(channels []slack.Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			slack.SlackResponse
			Channels         []slack.Channel `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}{
			SlackResponse: slack.SlackResponse{Ok: true},
			Channels:      channels,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func conversationsListFailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "internal_error",
		})
	}
}

func conversationsListPaginatedHandler() http.HandlerFunc {
	const page2Cursor = "page2"
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = r.ParseForm()
		cursor := r.Form.Get("cursor")
		resp := struct {
			slack.SlackResponse
			Channels         []slack.Channel `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}{
			SlackResponse: slack.SlackResponse{Ok: true},
		}
		if cursor == page2Cursor {
			resp.Channels = slackTestChannelsPage2
			resp.ResponseMetadata.NextCursor = ""
		} else {
			resp.Channels = slackTestChannelsPage1
			resp.ResponseMetadata.NextCursor = page2Cursor
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// typeListContains reports whether comma-separated Slack types includes want.
func typeListContains(typesCSV, want string) bool {
	for _, p := range strings.Split(typesCSV, ",") {
		if strings.TrimSpace(p) == want {
			return true
		}
	}
	return false
}

func encodeMissingScope(w http.ResponseWriter, needed string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      false,
		"error":   "missing_scope",
		"needed":  needed,
		"provided": "identify,channels:read",
	})
}

// conversationsListPrivateMissingScopeHandler fails any request that asks for
// private_channel (alone or combined). public_channel-only returns public fixtures.
// Combined requests fail so pre-per-type implementations hard-fail (RED for soft leaves).
func conversationsListPrivateMissingScopeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		types := r.Form.Get("types")
		// Empty types: Slack defaults to public+private; treat as multi-type fail path.
		if types == "" || typeListContains(types, "private_channel") {
			encodeMissingScope(w, "groups:read")
			return
		}
		// public_channel (and/or other non-private types): return public channels only.
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			slack.SlackResponse
			Channels         []slack.Channel `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}{
			SlackResponse: slack.SlackResponse{Ok: true},
			Channels:      slackTestPublicChannels,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func conversationsListAllMissingScopeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		types := r.Form.Get("types")
		needed := "groups:read"
		if typeListContains(types, "public_channel") && !typeListContains(types, "private_channel") {
			needed = "channels:read"
		}
		encodeMissingScope(w, needed)
	}
}

func limitFromRequest(r *http.Request) int {
	_ = r.ParseForm()
	limitStr := r.Form.Get("limit")
	if limitStr == "" {
		return 0
	}
	n, err := strconv.Atoi(limitStr)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func applyLimitNewestFirst(msgs []map[string]any, limit int) []map[string]any {
	if limit <= 0 || limit >= len(msgs) {
		return msgs
	}
	// msgs are newest-first; take the first `limit` (newest N).
	return msgs[:limit]
}

func historyHandler(fail bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if fail {
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "internal_error"})
			return
		}
		msgs := applyLimitNewestFirst(historyMessagesNewestFirst, limitFromRequest(r))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"messages": msgs,
			"has_more": false,
		})
	}
}

func repliesHandler(fail bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if fail {
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "internal_error"})
			return
		}
		msgs := applyLimitNewestFirst(threadRepliesNewestFirst, limitFromRequest(r))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"messages": msgs,
			"has_more": false,
		})
	}
}

// authTestSuccessHandler returns a stable auth.test body including bot_id for
// auth status leaves (registered before slacktest default /auth.test).
func authTestSuccessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"url":     slackTestAuthURL,
			"team":    slackTestTeamName,
			"user":    slackTestUserName,
			"team_id": slackTestTeamID,
			"user_id": slackTestUserID,
			"bot_id":  slackTestAuthBotID,
		})
	}
}

func authTestFailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "invalid_auth",
		})
	}
}

func startSlackTestServer(mode slackTestMode) (string, error) {
	sts := slacktest.NewTestServer(func(s slacktest.Customize) {
		switch mode {
		case slackTestChannelsFail:
			s.Handle("/conversations.list", conversationsListFailHandler())
		case slackTestChannelsPaginated:
			s.Handle("/conversations.list", conversationsListPaginatedHandler())
		case slackTestChannelsPrivateMissingScope:
			s.Handle("/conversations.list", conversationsListPrivateMissingScopeHandler())
		case slackTestChannelsAllMissingScope:
			s.Handle("/conversations.list", conversationsListAllMissingScopeHandler())
		default:
			s.Handle("/conversations.list", conversationsListHandler(slackTestChannels))
		}
		s.Handle("/conversations.history", historyHandler(mode == slackTestHistoryFail))
		s.Handle("/conversations.replies", repliesHandler(mode == slackTestHistoryFail))
		if mode == slackTestAuthFail {
			s.Handle("/auth.test", authTestFailHandler())
		} else {
			// Override slacktest default so unit auth status locks bot_id.
			s.Handle("/auth.test", authTestSuccessHandler())
		}
		if mode == slackTestPostFail {
			s.Handle("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"ok":    false,
					"error": "invalid_auth",
				})
			})
		}
	})
	sts.Start()
	return sts.GetAPIURL(), nil
}

func ensureSlackTestServer(t *testing.T) (string, error) {
	t.Helper()
	slackTestOnce.Do(func() {
		url, err := startSlackTestServer(slackTestDefault)
		if err != nil {
			slackTestErr = err
			return
		}
		slackTestURL = url
	})
	return slackTestURL, slackTestErr
}

func ensureSlackTestServerPostFail(t *testing.T) (string, error) {
	t.Helper()
	slackTestPostFailOnce.Do(func() {
		url, err := startSlackTestServer(slackTestPostFail)
		if err != nil {
			slackTestPostFailErr = err
			return
		}
		slackTestPostFailURL = url
	})
	return slackTestPostFailURL, slackTestPostFailErr
}

func ensureSlackTestServerHistoryFail(t *testing.T) (string, error) {
	t.Helper()
	slackTestHistoryFailOnce.Do(func() {
		url, err := startSlackTestServer(slackTestHistoryFail)
		if err != nil {
			slackTestHistoryFailErr = err
			return
		}
		slackTestHistoryFailURL = url
	})
	return slackTestHistoryFailURL, slackTestHistoryFailErr
}

func ensureSlackTestServerAuthFail(t *testing.T) (string, error) {
	t.Helper()
	slackTestAuthFailOnce.Do(func() {
		url, err := startSlackTestServer(slackTestAuthFail)
		if err != nil {
			slackTestAuthFailErr = err
			return
		}
		slackTestAuthFailURL = url
	})
	return slackTestAuthFailURL, slackTestAuthFailErr
}

func ensureSlackTestServerChannelsFail(t *testing.T) (string, error) {
	t.Helper()
	slackTestChannelsFailOnce.Do(func() {
		url, err := startSlackTestServer(slackTestChannelsFail)
		if err != nil {
			slackTestChannelsFailErr = err
			return
		}
		slackTestChannelsFailURL = url
	})
	return slackTestChannelsFailURL, slackTestChannelsFailErr
}

func ensureSlackTestServerChannelsPaginated(t *testing.T) (string, error) {
	t.Helper()
	slackTestChannelsPaginatedOnce.Do(func() {
		url, err := startSlackTestServer(slackTestChannelsPaginated)
		if err != nil {
			slackTestChannelsPaginatedErr = err
			return
		}
		slackTestChannelsPaginatedURL = url
	})
	return slackTestChannelsPaginatedURL, slackTestChannelsPaginatedErr
}

func ensureSlackTestServerChannelsPrivateMissingScope(t *testing.T) (string, error) {
	t.Helper()
	slackTestChannelsPrivateMissingScopeOnce.Do(func() {
		url, err := startSlackTestServer(slackTestChannelsPrivateMissingScope)
		if err != nil {
			slackTestChannelsPrivateMissingScopeErr = err
			return
		}
		slackTestChannelsPrivateMissingScopeURL = url
	})
	return slackTestChannelsPrivateMissingScopeURL, slackTestChannelsPrivateMissingScopeErr
}

func ensureSlackTestServerChannelsAllMissingScope(t *testing.T) (string, error) {
	t.Helper()
	slackTestChannelsAllMissingScopeOnce.Do(func() {
		url, err := startSlackTestServer(slackTestChannelsAllMissingScope)
		if err != nil {
			slackTestChannelsAllMissingScopeErr = err
			return
		}
		slackTestChannelsAllMissingScopeURL = url
	})
	return slackTestChannelsAllMissingScopeURL, slackTestChannelsAllMissingScopeErr
}

// mockAgentScript logs one flattened INVOCATION line (newlines in prompt → spaces)
// so multi-line open-inject prompts still match as a single agent launch record.
func mockAgentScript(logPath, successBody string, fail bool) string {
	// Interactive open (agentrunbridge WaitReady) polls: <binary> tty status <sessionID>.
	// Ready fixture must not count toward WantAgentCalls launch lines.
	tail := fmt.Sprintf("printf %%s %q\n", successBody)
	if fail {
		tail = "echo \"mock agent failed\" >&2\nexit 1\n"
	}
	return fmt.Sprintf(`#!/bin/sh
if [ "$1" = "tty" ] && [ "$2" = "status" ]; then
  printf '%%s\n' 'screen status: banner'
  printf '%%s\n' 'sendable: yes'
  exit 0
fi
line="INVOCATION"
for a in "$@"; do
  flat=$(printf '%%s' "$a" | tr '\n' ' ')
  line="$line $flat"
done
printf '%%s\n' "$line" >> %q
%s`, logPath, tail)
}

func writeMockAgent(t *testing.T, dir, logPath string) string {
	t.Helper()
	path := filepath.Join(dir, "mock-agent-run")
	if err := os.WriteFile(path, []byte(mockAgentScript(logPath, defaultAgentReply, false)), 0o755); err != nil {
		t.Fatalf("write mock agent: %v", err)
	}
	return path
}

// writeMockAgentFail logs the launch line then exits non-zero (agent open error path).
func writeMockAgentFail(t *testing.T, dir, logPath string) string {
	t.Helper()
	path := filepath.Join(dir, "mock-agent-run-fail")
	if err := os.WriteFile(path, []byte(mockAgentScript(logPath, "", true)), 0o755); err != nil {
		t.Fatalf("write mock agent fail: %v", err)
	}
	return path
}

func expectedDefaultLockPath(homeDir string) string {
	return filepath.Join(homeDir, filepath.FromSlash(defaultListenLockRelPath))
}

func expectedSessionSystemMDPath(homeDir, sessionID string) string {
	return filepath.Join(homeDir, filepath.FromSlash(defaultSlackLocalBotRelDir), "sessions", sessionID, "SYSTEM.md")
}

func expectedSessionsJSONPath(homeDir string) string {
	return filepath.Join(homeDir, filepath.FromSlash(defaultSlackLocalBotRelDir), "sessions.json")
}

func expectedMessagesJSONLPath(homeDir, sessionID string) string {
	return filepath.Join(homeDir, filepath.FromSlash(defaultSlackLocalBotRelDir), "sessions", sessionID, "messages.jsonl")
}

// sessionMapEntry is a durable sessions.json entry (subset used by leaves).
type sessionMapEntry struct {
	SessionID          string `json:"session_id"`
	ChannelID          string `json:"channel_id"`
	ThreadTS           string `json:"thread_ts"`
	ConfigPath         string `json:"config_path"`
	Dir                string `json:"dir,omitempty"` // optional agent workspace (abs preferred)
	Kind               string `json:"kind"`
	ReplyMode          string `json:"reply_mode"`
	LastMessagePreview string `json:"last_message_preview,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

type sessionsJSONFile struct {
	Version int               `json:"version"`
	Entries []sessionMapEntry `json:"entries"`
}

type sessionLogMessage struct {
	MessageID string `json:"message_id"`
	TS        string `json:"ts"`
	User      string `json:"user"`
	Text      string `json:"text"`
	Direction string `json:"direction"` // in | out
}

func seedSessionsJSON(t *testing.T, homeDir string, entries []sessionMapEntry) error {
	t.Helper()
	path := expectedSessionsJSONPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if entries == nil {
		entries = []sessionMapEntry{}
	}
	for i := range entries {
		if entries[i].Kind == "" {
			entries[i].Kind = "channel"
		}
		if entries[i].ReplyMode == "" {
			entries[i].ReplyMode = "channel"
		}
	}
	data, err := json.MarshalIndent(sessionsJSONFile{Version: 1, Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func seedMessagesJSONL(t *testing.T, homeDir, sessionID string, msgs []sessionLogMessage) error {
	t.Helper()
	path := expectedMessagesJSONLPath(homeDir, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, m := range msgs {
		line, err := json.Marshal(m)
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func readMessagesJSONL(t *testing.T, homeDir, sessionID string) ([]sessionLogMessage, error) {
	t.Helper()
	path := expectedMessagesJSONLPath(homeDir, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []sessionLogMessage
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m sessionLogMessage
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			return nil, fmt.Errorf("parse messages.jsonl line %q: %w", line, err)
		}
		out = append(out, m)
	}
	return out, nil
}

func readSessionsJSON(t *testing.T, homeDir string) (*sessionsJSONFile, error) {
	t.Helper()
	path := expectedSessionsJSONPath(homeDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc sessionsJSONFile
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func isolateHome(t *testing.T, req *Request) error {
	t.Helper()
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if req.HomeDir == "" {
		req.HomeDir = filepath.Join(req.WorkDir, "home")
	}
	return os.MkdirAll(req.HomeDir, 0o755)
}

func newListenSlackTestServer(t *testing.T, posts *[]CapturedPost) (*slacktest.Server, string) {
	t.Helper()
	sts := slacktest.NewTestServer(func(s slacktest.Customize) {
		s.Handle("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
			data, _ := io.ReadAll(r.Body)
			values, _ := url.ParseQuery(string(data))
			if posts != nil {
				*posts = append(*posts, CapturedPost{
					Channel:  values.Get("channel"),
					Text:     values.Get("text"),
					ThreadTS: values.Get("thread_ts"),
				})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":      true,
				"channel": values.Get("channel"),
				"ts":      fmt.Sprintf("%d.000001", time.Now().Unix()),
				"text":    values.Get("text"),
			})
		})
	})
	sts.Start()
	t.Cleanup(sts.Stop)
	return sts, sts.GetAPIURL()
}

func socketModeEnvelope(envelopeID string, eventType string, inner map[string]any) (string, error) {
	inner["type"] = eventType
	payload := map[string]any{
		"token":      slackTestBotToken,
		"team_id":    slackTestTeamID,
		"type":       "event_callback",
		"event":      inner,
		"event_id":   "Ev" + envelopeID,
		"event_time": time.Now().Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	envelope := map[string]any{
		"envelope_id":              envelopeID,
		"payload":                  json.RawMessage(payloadBytes),
		"type":                     "events_api",
		"accepts_response_payload": false,
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func injectEvent(sts *slacktest.Server, ev InjectedEvent) error {
	channel := ev.Channel
	user := ev.User
	if user == "" {
		user = slackTestUserID
	}
	ts := ev.TS
	if ts == "" {
		ts = fmt.Sprintf("%d.000100", time.Now().Unix())
	}
	kind := ev.Kind
	if kind == "" {
		kind = "message"
	}
	if kind == "dm" {
		channel = slackTestDMChannelID
		kind = "message"
	}
	inner := map[string]any{
		"user":    user,
		"text":    ev.Text,
		"ts":      ts,
		"channel": channel,
	}
	if ev.ThreadTS != "" {
		inner["thread_ts"] = ev.ThreadTS
	}
	eventType := "message"
	if kind == "app_mention" {
		eventType = "app_mention"
	}
	// Unique envelope ids even when dual events share channel+ts (dedupe leaf).
	envelopeID := fmt.Sprintf("Env-%s-%s-%d", eventType, strings.ReplaceAll(ts, ".", ""), time.Now().UnixNano())
	msg, err := socketModeEnvelope(envelopeID, eventType, inner)
	if err != nil {
		return err
	}
	sts.SendToWebsocket(msg)
	return nil
}

func readAgentInvocations(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Launch lines only; ignore any non-INVOCATION noise (status polls must not log).
		if strings.HasPrefix(line, "INVOCATION") {
			out = append(out, line)
		}
	}
	return out, nil
}

func waitForAgentLog(path string, wantMin int, timeout time.Duration) ([]string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		lines, err := readAgentInvocations(path)
		if err != nil {
			return nil, err
		}
		if len(lines) >= wantMin {
			return lines, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	lines, _ := readAgentInvocations(path)
	return lines, fmt.Errorf("timeout waiting for agent log %s (got %d, want %d)", path, len(lines), wantMin)
}

// safeBuffer is a bytes.Buffer usable concurrently from pipe copy + readiness polls.
type safeBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func waitForStdoutContains(buf *safeBuffer, needle string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), needle) {
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %q in listen stdout", needle)
}

func waitForPosts(posts *[]CapturedPost, wantMin int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if posts != nil && len(*posts) >= wantMin {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	n := 0
	if posts != nil {
		n = len(*posts)
	}
	return fmt.Errorf("timeout waiting for post messages (got %d, want %d)", n, wantMin)
}

func isSubcommand(s string) bool {
	switch s {
	case "send", "history", "listen", "channels", "auth", "session":
		return true
	default:
		return false
	}
}

func insertConfigAfterSubcommand(args []string, configPath string) []string {
	cfg := []string{"--config", configPath}
	if len(args) > 0 && isSubcommand(args[0]) {
		// channels list|search, auth status, session reply|history: place --config after action.
		if args[0] == "channels" && len(args) > 1 && (args[1] == "list" || args[1] == "search") {
			out := make([]string, 0, len(args)+2)
			out = append(out, args[0], args[1])
			out = append(out, cfg...)
			out = append(out, args[2:]...)
			return out
		}
		if args[0] == "auth" && len(args) > 1 && args[1] == "status" {
			out := make([]string, 0, len(args)+2)
			out = append(out, args[0], args[1])
			out = append(out, cfg...)
			out = append(out, args[2:]...)
			return out
		}
		if args[0] == "session" && len(args) > 1 && (args[1] == "reply" || args[1] == "history") {
			out := make([]string, 0, len(args)+2)
			out = append(out, args[0], args[1])
			out = append(out, cfg...)
			out = append(out, args[2:]...)
			return out
		}
		out := make([]string, 0, len(args)+2)
		out = append(out, args[0])
		out = append(out, cfg...)
		out = append(out, args[1:]...)
		return out
	}
	return append(cfg, args...)
}

func defaultListenArgs(req *Request) []string {
	args := []string{"listen"}
	if req.BotToken != "" {
		args = append(args, "--token", req.BotToken)
	}
	if req.AppToken != "" {
		args = append(args, "--app-token", req.AppToken)
	}
	if req.ConfigPath != "" {
		args = append(args, "--config", req.ConfigPath)
	}
	if req.NoLock {
		args = append(args, "--no-lock")
	} else if req.LockFile != "" {
		args = append(args, "--lock-file", req.LockFile)
	}
	// UseDefaultLock: omit both flags so product default path applies.
	args = append(args, req.Args...)
	return args
}

func runSimple(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if req.HomeDir != "" {
		if err := os.MkdirAll(req.HomeDir, 0o755); err != nil {
			return nil, err
		}
	}
	if err := materializeConfig(t, d, req); err != nil {
		return nil, err
	}
	var capturedPosts []CapturedPost
	if req.CapturePosts {
		_, apiURL := newListenSlackTestServer(t, &capturedPosts)
		req.SlackAPIURL = apiURL
	}
	if req.HistoryAPIFail {
		apiURL, err := ensureSlackTestServerHistoryFail(t)
		if err != nil {
			return nil, err
		}
		req.SlackAPIURL = apiURL
	}
	if req.AuthAPIFail {
		apiURL, err := ensureSlackTestServerAuthFail(t)
		if err != nil {
			return nil, err
		}
		req.SlackAPIURL = apiURL
	}
	if req.ChannelsAPIFail {
		apiURL, err := ensureSlackTestServerChannelsFail(t)
		if err != nil {
			return nil, err
		}
		req.SlackAPIURL = apiURL
	}
	if req.ChannelsPaginated {
		apiURL, err := ensureSlackTestServerChannelsPaginated(t)
		if err != nil {
			return nil, err
		}
		req.SlackAPIURL = apiURL
	}
	if req.ChannelsPrivateMissingScope {
		apiURL, err := ensureSlackTestServerChannelsPrivateMissingScope(t)
		if err != nil {
			return nil, err
		}
		req.SlackAPIURL = apiURL
	}
	if req.ChannelsAllMissingScope {
		apiURL, err := ensureSlackTestServerChannelsAllMissingScope(t)
		if err != nil {
			return nil, err
		}
		req.SlackAPIURL = apiURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	cmd.Dir = req.WorkDir

	env := os.Environ()
	if req.ClearSlackEnv {
		env = withoutEnvKeys(env,
			"SLACK_BOT_TOKEN", "SLACK_APP_TOKEN", "SLACK_CHANNEL", "SLACK_CONFIG", "SLACK_API_URL",
			"SLACK_MSG_SESSION_ID", "SLACK_MSG_CONFIG",
			envAgentRun, envAgentLog,
		)
	}
	if req.HomeDir != "" {
		env = withoutEnvKeys(env, "HOME", "USERPROFILE")
		env = append(env, "HOME="+req.HomeDir)
	}
	if req.SlackAPIURL != "" {
		env = withoutEnvKeys(env, "SLACK_API_URL")
		env = append(env, "SLACK_API_URL="+req.SlackAPIURL)
	}
	env = mergeEnv(env, req.Env...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout:       stdout.String(),
		Stderr:       stderr.String(),
		PostMessages: capturedPosts,
	}
	if err == nil {
		resp.ExitCode = 0
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runListenQuick(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if err := materializeConfig(t, d, req); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, defaultListenArgs(req)...)
	cmd.Dir = req.WorkDir
	env := os.Environ()
	if req.ClearSlackEnv {
		env = withoutEnvKeys(env,
			"SLACK_BOT_TOKEN", "SLACK_APP_TOKEN", "SLACK_CONFIG", "SLACK_API_URL",
			"SLACK_MSG_SESSION_ID", "SLACK_MSG_CONFIG",
			envAgentRun, envAgentLog,
		)
	}
	if req.HomeDir != "" {
		env = withoutEnvKeys(env, "HOME", "USERPROFILE")
		env = append(env, "HOME="+req.HomeDir)
	}
	if req.SlackAPIURL != "" {
		env = withoutEnvKeys(env, "SLACK_API_URL")
		env = append(env, "SLACK_API_URL="+req.SlackAPIURL)
	}
	if req.MockAgentPath != "" {
		env = mergeEnv(env, envAgentRun+"="+req.MockAgentPath)
	}
	if req.AgentLogPath != "" {
		env = mergeEnv(env, envAgentLog+"="+req.AgentLogPath)
	}
	env = mergeEnv(env, req.Env...)
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{Stdout: stdout.String(), Stderr: stderr.String()}
	if err == nil {
		resp.ExitCode = 0
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runDaemon(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if err := materializeConfig(t, d, req); err != nil {
		return nil, err
	}
	if req.HomeDir != "" {
		if err := os.MkdirAll(req.HomeDir, 0o755); err != nil {
			return nil, err
		}
	}
	// Isolate product default lock across parallel leaves unless leaf opts in.
	if !req.NoLock && req.LockFile == "" && !req.UseDefaultLock {
		req.LockFile = filepath.Join(req.WorkDir, "slack-msg.listen.lock")
	}
	if req.AgentLogPath == "" {
		req.AgentLogPath = filepath.Join(req.WorkDir, "agent.log")
	}
	if req.MockAgentPath == "" {
		req.MockAgentPath = writeMockAgent(t, req.WorkDir, req.AgentLogPath)
	}
	var posts []CapturedPost
	if req.Posts != nil {
		posts = *req.Posts
	}
	sts, apiURL := newListenSlackTestServer(t, &posts)
	req.SlackAPIURL = apiURL
	timeout := req.ObserveTimeout
	if timeout == 0 {
		timeout = 8 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, defaultListenArgs(req)...)
	cmd.Dir = req.WorkDir
	env := os.Environ()
	if req.ClearSlackEnv {
		env = withoutEnvKeys(env,
			"SLACK_BOT_TOKEN", "SLACK_APP_TOKEN", "SLACK_CONFIG", "SLACK_API_URL",
			"SLACK_MSG_SESSION_ID", "SLACK_MSG_CONFIG",
			envAgentRun, envAgentLog,
		)
	}
	if req.HomeDir != "" {
		env = withoutEnvKeys(env, "HOME", "USERPROFILE")
		env = append(env, "HOME="+req.HomeDir)
	}
	env = mergeEnv(env,
		"SLACK_API_URL="+req.SlackAPIURL,
		envAgentRun+"="+req.MockAgentPath,
		envAgentLog+"="+req.AgentLogPath,
	)
	env = mergeEnv(env, req.Env...)
	cmd.Env = env
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var stdoutBuf, stderrBuf safeBuffer
	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _, _ = io.Copy(&stdoutBuf, stdoutPipe) }()
		go func() { defer wg.Done(); _, _ = io.Copy(&stderrBuf, stderrPipe) }()
		wg.Wait()
		close(done)
	}()
	// Wait for startup banner (auth done) before inject. Fixed 500ms slept under
	// full repo-l2 parallel load and left agent.log empty (got 0).
	bannerWait := timeout
	if bannerWait > 5*time.Second {
		bannerWait = 5 * time.Second
	}
	if bannerWait < 2*time.Second {
		bannerWait = 2 * time.Second
	}
	if err := waitForStdoutContains(&stdoutBuf, "Using config from:", bannerWait); err != nil {
		_ = cmd.Process.Kill()
		<-done
		return nil, fmt.Errorf("%w\nstdout:\n%s\nstderr:\n%s", err, stdoutBuf.String(), stderrBuf.String())
	}
	// Socket Mode WS connects after the banner; brief settle under CI load.
	time.Sleep(250 * time.Millisecond)
	for i, ev := range req.InjectEvents {
		if err := injectEvent(sts, ev); err != nil {
			_ = cmd.Process.Kill()
			<-done
			return nil, err
		}
		if i < len(req.InjectEvents)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	wantAgent := req.WantAgentCalls
	if wantAgent < 0 {
		wantAgent = len(req.InjectEvents)
	}
	var invocations []string
	var agentErr error
	if wantAgent > 0 {
		invocations, agentErr = waitForAgentLog(req.AgentLogPath, wantAgent, timeout)
	} else {
		time.Sleep(800 * time.Millisecond)
		invocations, _ = readAgentInvocations(req.AgentLogPath)
	}
	if req.SecondInstance {
		secondCtx, secondCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer secondCancel()
		second := exec.CommandContext(secondCtx, req.Bin, defaultListenArgs(req)...)
		second.Dir = req.WorkDir
		second.Env = env
		var sOut, sErr bytes.Buffer
		second.Stdout = &sOut
		second.Stderr = &sErr
		err := second.Run()
		resp := &Response{
			Stdout:           stdoutBuf.String(),
			Stderr:           stderrBuf.String(),
			AgentInvocations: invocations,
			PostMessages:     posts,
			SecondStdout:     sOut.String(),
			SecondStderr:     sErr.String(),
		}
		if err == nil {
			resp.SecondExitCode = 0
		} else {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				resp.SecondExitCode = exitErr.ExitCode()
			}
		}
		_ = cmd.Process.Signal(syscall.SIGTERM)
		<-done
		return resp, nil
	}
	if wantAgent > 0 && agentErr != nil {
		_ = cmd.Process.Kill()
		<-done
		return &Response{Stdout: stdoutBuf.String(), Stderr: stderrBuf.String(), AgentInvocations: invocations, PostMessages: posts}, agentErr
	}
	// Thread interactive open does not PostMessage agent body; only wait when leaf asks.
	if req.WantPosts > 0 {
		_ = waitForPosts(&posts, req.WantPosts, 3*time.Second)
	} else if wantAgent > 0 {
		// Brief settle so late captures (if any) land before SIGTERM; no full timeout.
		time.Sleep(150 * time.Millisecond)
	}
	_ = cmd.Process.Signal(syscall.SIGTERM)
	<-done
	_ = cmd.Wait()
	return &Response{
		Stdout:           stdoutBuf.String(),
		Stderr:           stderrBuf.String(),
		AgentInvocations: invocations,
		PostMessages:     posts,
	}, nil
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.Bin == "" {
		return nil, fmt.Errorf("req.Bin not set; root Setup must build slack-msg")
	}
	if req.ListenMode || req.Daemon || req.SecondInstance {
		if req.BotToken == "" && !req.ClearSlackEnv {
			req.BotToken = slackTestBotToken
		}
		if req.AppToken == "" && !req.ClearSlackEnv && req.Daemon {
			req.AppToken = slackTestAppToken
		}
		if req.Daemon || req.SecondInstance {
			return runDaemon(t, d, req)
		}
		return runListenQuick(t, d, req)
	}
	return runSimple(t, d, req)
}

func _dsnSlackEventsUnused() {
	_ = slackevents.AppMentionEvent{}
}
```
