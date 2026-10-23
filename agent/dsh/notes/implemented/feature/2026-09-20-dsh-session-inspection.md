# Agent Note: Read-only DSH session inspection

Status: implemented

## Problem

Agent-pro's provider commands expose stored session history, but DSH uses a different, versioned artifact format with compressed logs by default. A file's last turn outcome is not evidence that its owning process is live, and launching a headless DSH process is not equivalent to sending input to an existing Web session.

## Decision

The DSH inspector reads v3 and v4 JSONL and Zstandard artifacts directly in Go. It selects the highest canonical generation, retains original event records for NDJSON, and presents recorded history rather than reconstructing DSH's current model context. It writes no session state and does not acquire writer ownership. Unsupported generations remain explicit failures; older readable predecessors are not fallbacks. Cross-session listings isolate damaged or unsupported artifacts with diagnostics, while exact inspection fails.

## Alternatives considered

**A DSH-owned inspection API:** This would reuse DSH's codecs and migration semantics, but requires shipping an additional DSH interface before agent-pro can inspect local sessions. The native inspector instead accepts an explicit v3-and-v4 physical compatibility limit (same JSONL row framing; no DSH migration codec).

**Plain JSONL only:** This avoids a decompression dependency but excludes the default DSH storage configuration. A maintained Go Zstandard decoder supports ordinary installations without requiring a separate executable.

**Headless continuation as session send:** Headless DSH requires compatible cwd and composition, refuses presets and child/fork sessions, and selects its invoking profile's default model. It is not a faithful substitute for the existing live session owner. Live control remains absent from these commands.

## Consequences

Session inspection works without a DSH executable, server, or provider credentials. Agent-pro owns a deliberately limited reader that must be reviewed when DSH changes its physical format. Unknown plugin records remain available in raw output but are not interpreted as executable behavior. Tests cover generation selection, encoding, projections, diagnostics, command output, and unchanged source artifacts; they do not establish parity with DSH's complete restoration engine.
