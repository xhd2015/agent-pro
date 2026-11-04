# DSH Web sessions as a Go library

Import `github.com/xhd2015/agent-pro/agent/dsh/web` (usually aliased as `dshweb`).
This package does **not** execute `agent-run`, `dsh`, iTerm, a shell script, or a
TTY readiness probe. It talks to an already-running DSH Web service and uses the
OS browser opener. Optional model/reasoning overrides initialize only the new
session; they never save DSH global defaults.

## High-level lifecycle

```go
service := dshweb.Service{
    Client: dshweb.Client{}, // DSH_HOME/web-listen.json, or ~/.dsh
    Store: dshweb.FileStore{
        Dir: "/application/chosen/session-state",
        FileName: "runner-binding.json", // optional; default is session.json
    },
}

result, err := service.OpenSession(ctx, dshweb.SessionOptions{
    Key: "business-conversation-123",
    Workspace: "/workspace",
    Browser: "brave", // empty = system default
    Prompt: "Read the playbook and help with this task.",
})
// Handle err first. result.Accepted means input/draft admission, not task completion.
// Present result.Warnings separately: a browser failure after input admission
// must NOT cause the application to redeliver its message.
_ = result
_ = err

result, err = service.Send(ctx, dshweb.SendOptions{
    Key: "business-conversation-123",
    Prompt: "Here is the next message.",
})

// Focus / reopen: no creation and no prompt submission (including pending input).
result, err = service.Open(ctx, "business-conversation-123")
```

| Operation | Contract |
| --- | --- |
| `OpenSession` | Reserve a provider identity, create/reuse the session, admit the initial prompt once, and automatically open its browser once unless suppressed. Repeated calls preserve pinned workspace/browser/initial model and do not resubmit admitted initial input. An empty prompt permits a blank session. |
| `Send` | Require an existing binding, settle any interrupted input with its original request ID, then admit the next message. Never automatically pop a browser. |
| `Open` | Open an existing, created session only. Never create a session or submit/retry pending input. |

`NoSubmit` stores a composer draft instead of submitting a model turn. `NoOpen`
on `OpenSession` suppresses automatic browser opening for that call without affecting
admission. `Browser: BrowserNone` (`"none"`) instead pins automatic opening off for
the conversation, including retries and restarts. Creation, model selection and
prompt submission still run normally; successful admission returns `Accepted: true`
and `BrowserOpened: false`, without a browser warning. Do not use `NoSubmit` for
background execution. Empty/omitted browser still means the system default.

```go
result, err := service.OpenSession(ctx, dshweb.SessionOptions{
    Key: conversationID, Workspace: workspace, Prompt: prompt,
    Browser: dshweb.BrowserNone,
})
```

`Service.Open` is an explicit user action: a `BrowserNone` binding opens in the
system default browser without changing its pinned preference or delivering input.
Low-level `Client.Open(id, BrowserNone)` is a no-op and does not discover the server
or call an OS opener. Other supported browser names are brave, chrome, firefox,
edge, safari and opera; `NormalizeBrowser`/`IsRunner` are reusable validation helpers.
Background sessions still require a running DSH Web service and can wait for human
approval. This policy does not disable browser-automation tools used by the agent.

## Live models and Session-only overrides

`Client.ModelCatalog(ctx)` calls `session/modelCatalog` without creating a session,
opening a browser or changing configuration. It retains `Default`, provider
`Groups`, each model's `Reasoning.Efforts`/`DefaultEffort`, model `Unavailable`
reasons, and isolated provider `Failures`. Keep `(provider, model)` together: two
providers may advertise the same model id. A model with a non-empty `Unavailable`
reason must not be offered as selectable. Effort ids are adapter-owned strings,
not a fixed Codex enum. A partial failure does not erase the healthy groups.

Pass the selected values in `SessionOptions.ModelSelection`:

```go
options.ModelSelection = &dshweb.ModelSelection{
    Provider: providerID,
    Model: modelID,
    ReasoningEffort: effortID, // empty follows this model's adapter default
}
result, err := service.OpenSession(ctx, options)
```

A nil selection preserves the existing DSH-default behavior. A new binding saves
its initial selection before creation, and the Host's accepted response before
sending input. The optional `initial_model_selection` field uses the existing
binding file; no separate mapping is introduced. Restarts and creation retries
use that binding, not later application preferences. `Send`, pure `Open`, and
repeated `OpenSession` never reapply a new preference to an existing session.
The DSH Session log remains authoritative after a user changes its model in DSH.

`Client.CreateWithOptions` is the low-level equivalent; existing `Client.Create`
callers keep their unchanged default behavior. An explicit override requires a
Host that accepts and echoes `session/create.modelSelection`. A missing/mismatched
echo stops before prompt or browser delivery rather than silently using the
wrong model or calling the globally mutating `session/selectModel`. An older
Host may already have created a blank session; update it and retry with the same
reserved identity. Incomplete provider/model pairs fail before binding reservation;
DSH validates live model and effort availability before installing the choice.

## Ownership and recovery

The application owns business-to-key mapping, runner selection, prompt formatting,
preferences, and the state root. The service owns provider IDs, pending request
IDs/input, creation/admission checkpoints, browser-opening bookkeeping and retry
sequencing. Applications must not share one key between different conversations.

`FileStore` keeps `<Dir>/<Key>/<FileName>` and a sibling lock file. Writes use an
owner-only temporary file, file sync and atomic rename. The filesystem lock spans
all intermediate checkpoints and RPCs, serializing distinct service objects and
processes. Waiting for the lock respects context cancellation. No global
application home or package-global session registry is used.

The `Store` / `Record` interfaces also support application-owned persistence.
`WithSession` implementations must serialize one key and `Record.Save` must commit
an intermediate checkpoint immediately, not defer all writes until callback exit.
`Read` observes the last atomic checkpoint without waiting for a network operation.

Flat legacy Marcus records are compatible without conversion: `dsh_session_id`,
`created`, `initial_delivered`, `browser_opened`, and pending fields keep their
names. Unknown application metadata survives writes. Non-DSH runner pins and
corrupt records fail closed instead of being rebound to a new provider session.
Launch tokens are discovered per operation and never stored in session state.
Confirmed pending input is removed from the durable checkpoint.

A browser failure after admission yields `browser_open_failed` in `Warnings` and
no delivery error. A browser-bookkeeping write failure similarly yields
`browser_checkpoint_failed`. Pure `Open` failures return an error because there
was no input admission to distinguish. Transport/checkpoint failures leave the
last committed request ID available for retry. No error falls back to a terminal.

`Client` remains available for consumers that already own equivalent session
orchestration; prefer `Service` to avoid duplicating lifecycle and recovery logic.

## Verification

```sh
go test -race ./agent/dsh/web -count=3
go test ./agent/dsh/... ./pkgs/agentruncli
```

Coverage includes restart/reuse, pinned preferences, drafts, pure reopen,
ambiguous delivery deduplication, recovery before newer input, browser warnings,
failed checkpoint writes, legacy state, validation, lock cancellation, concurrent
instances, and a sparse multi-process test using test binaries (no real browser
or model invocation). The library prints no application-specific logs.
