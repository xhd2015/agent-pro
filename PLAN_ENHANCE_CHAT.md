# Plan: Enhance Chat UI for Pure Session Events (PLAN_ENHANCE_CHAT)

## Background

The web chat for `agent-run` (especially TTY runners like `grok-tty`) was incorrectly mixing terminal/PTY content into the chat message area. This happened because:

- Real grok-tty runs (non-mock) often failed session ID discovery / updates tailing.
- Fallback logic extracted text from PTY snapshots and emitted it as `assistant` messages.
- `KeepTerminalAlive` + no-phases path caused growing incremental message events.
- These polluted `events.jsonl`, which the web chat (and other consumers) rendered.

The original web-cli-subset refactor intended web to be a strict subset of CLI: same events, same storage. The chat should be a thin renderer of standardized session events.

## Goals (Confirmed Direction)

- The **chat area is exclusively** for streaming and rendering standardized session events (`agent/event/types` shapes from the session's events source / SSE).
- **Nothing** from the terminal/PTY may appear in chat messages (no snapshots, no extraction, no chrome, no echoed prompts, no "final answer from scrollback").
- There is a separate **terminal entrypoint** (button / link) for viewing PTY status, transcript, and live attach. Chat and terminal are fully decoupled.
- **All progress messages** (including binding steps) must be shown as cards in the chat area.
- Frontend is a pure event streamer + renderer. Streaming is event-driven only (no client-side growth or synthesis).
- Complete refactor: drop legacy terminal-binding logic from the chat flow. No preserving old behavior that injected terminal content into events or UI.
- For TTY runners, the chat timeline shows only what the session event log actually contains.

## Specific Requirements (from Clarification)

### Q1: Session ID / Log Binding
- During resolution: show progress "Resolve session id..." (as a card in chat).
- On error: show "Cannot resolve session id: <concrete error>" (exact error text surfaced).
- These are visible in the frontend chat area.

### Q2: Progress Messages
- Show **all kinds of progress messages** as cards.
- This includes binding, buffering, thinking steps, tool phases, etc.

### Q3: Streaming Model
- Yes: purely event-driven.
- Frontend streams what the session events have (convert/render to standard types) and displays them.
- No client-side growth, animation, or terminal-derived simulation for the chat timeline.

### Q4: Scope
- Yes: complete refactor.
- No existing terminal binding in the chat path.
- Chat and terminal are separate surfaces.

## Target UI Behavior (Conceptual)

### Normal Flow (Successful Binding + Events)
- Chat shows user message.
- Progress cards appear for "Resolve session id...".
- Real events stream in as cards/messages: thoughts, tool calls, assistant replies, done.
- Terminal button remains available at all times (separate from chat content).
- No terminal transcript ever leaks into the message list.

### Error Flow (Binding Failure)
- User message is shown.
- Error card: "Cannot resolve session id: <concrete error>".
- Chat area contains only events + the explicit error. No fallback content.
- User can still click the terminal entrypoint to inspect the PTY directly.

### Progress Cards
- Binding, "buffering", intermediate states, etc. all appear uniformly as cards driven by events.
- Cards are part of the event stream, not special-cased outside events.

### Separation
- Chat area = event renderer only.
- Terminal button/modal = PTY view (status, attach, full transcript if needed).
- When no structured events exist for a turn, the chat stays minimal (user message + status/error cards). Direct user to terminal button.

## High-Level Approach

### 1. Producer Side (Event Emission for TTY Runners)
- Only write standardized, protocol-derived events into the session event log.
- Remove any code that turns PTY snapshots, scrollback, or terminal content into `ActionMessage` (assistant) or similar events.
- Session ID resolution / discovery must emit proper progress events (e.g. think or dedicated progress type) so the frontend can show "Resolve session id..." as a card.
- On resolution failure, emit a clear error event containing the concrete message.
- KeepTerminalAlive continues to be used for PTY lifetime/attach/send semantics, but it must not affect what is written to the event log for chat.
- Improve reliability of discovery / sid capture for real (non-harness) grok-tty launches so the clean tail path is taken more often.
- When the tail cannot attach, do **not** synthesize anything from the terminal. The event log simply lacks assistant content for that turn.

### 2. Frontend / Chat Renderer
- Chat becomes a thin consumer of the existing session events stream (SSE / file tail).
- Render standardized events directly (user, assistant, think, tool, done, progress, error).
- Progress and error messages (including binding) are rendered as cards.
- Remove any client-side logic that was extracting, growing, or synthesizing content from terminal sources.
- Streaming is purely event-driven: new events appear as they arrive.
- Keep the terminal button as an independent UI element that opens the separate PTY surface.
- For TTY runners the chat timeline may contain binding progress cards followed by real events (or error cards if binding failed). No terminal-derived text.

### 3. Event Types & Progress
- Use (or introduce) event shapes that support progress cards and concrete errors.
- Binding step ("Resolve session id...") and its error variant must be representable so the renderer can display them appropriately.
- All progress (not just binding) should flow through the same card rendering path.

### 4. Complete Refactor Scope
- Drop legacy paths that tied chat content to terminal state.
- Update web run options, agentui emission, TTY runner tail logic, and any web-specific handlers.
- Frontend chat component(s) must enforce "events only".
- Existing doctest harnesses (which already produce clean events) continue to work; add coverage for real-binary binding success + error cases.

## Non-Goals / Constraints
- Do not put terminal transcript or snapshot content into chat messages under any circumstance.
- Chat must not depend on the terminal attach/PTY view for its content.
- Preserve (or improve) CLI parity for events produced by grok-tty runs.
- The terminal surface itself remains available and unchanged for users who need the raw PTY.

## Verification Criteria
- A real `agent-run web --agent-runner grok-tty` session shows only proper events (or binding progress / error cards) in the chat area.
- Clicking the terminal button opens a separate view with no impact on the chat message list.
- "Resolve session id..." and concrete error messages appear as expected.
- All progress appears as cards.
- No terminal chrome, boxes, echoed prompts, or snapshot text ever appears in chat.
- Pure event streaming works; no client-side growth from terminal data.
- Behavior is consistent whether the binding succeeds or fails.
- Existing CLI consumers of events are unaffected or improved.

## Open Questions (for further clarification if needed)
- Exact event shape for "Resolve session id..." progress and its error (new type? reuse think/error?).
- Visual treatment of cards vs. normal messages (any distinction beyond content?).
- What (if anything) should appear in chat when a TTY turn produces zero protocol events (beyond the explicit error card on binding failure)?
- Any special handling for the very first "Resolve session id..." card vs. later progress?

This plan is derived directly from the investigation findings and the clarified requirements in the followup discussion. It treats the chat area as a pure event renderer and fully decouples it from terminal/PTY concerns.