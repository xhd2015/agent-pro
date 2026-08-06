---
name: verify-on-behalf-of-user/transcript
description: >-
  Markdown transcript format rules, template usage, write-to-file plus inline
  full content for direct review, and quick-start commands. Header must include
  Mode (sandbox|host) and labeled depth.
---

# Transcript

The primary deliverable is a markdown file **and** that same content inlined in
the agent reply so the user can review without opening the file.

## Write steps

1. Start from `templates/transcript.md` (installed with the skill)
2. Fill header table (**Mode** `sandbox`\|`host`, scope, **depth + reason**, surface, verdict, paths)
3. Append each phase as fenced ` ```sh ` blocks with `$ ` prompts
4. Insert `> **Annotation:**` / `> **Check:** ✓|✗` between blocks
5. Save to `~/.sandbox/transcripts/<ISO8601>-<slug>.md` (even when Mode is `host` —
   path is for logs only; Mode field records the runtime)
6. **Read the file back and put its full body in the agent reply** (verbatim)
7. Lead the reply with path, **Mode**, labeled depth, and one-line verdict

## Agent final reply shape

```text
## Verdict
**PASS** | **FAIL**

## Mode
Mode: sandbox | host

## Depth
Depth: scenario
Reason: server + UI claim requires real bring-up and browser-agent

## Transcript path
`~/.sandbox/transcripts/2026-07-16T120000-my-feature.md`

## Transcript
<entire file contents — same bytes as on disk>
```

Do **not** summarize-only. If the channel truncates extremely large dumps, note
truncation and keep the file as source of truth — default is **full** content.

## Transcript format rules

| Element | Format |
|---------|--------|
| Mode | `sandbox` (default) or `host` (explicit opt-in only) — required in header |
| Depth | `Depth: smoke\|scenario\|full` + `Reason:` in header (required) |
| Surface | e.g. `CLI-only`, `server+frontend` |
| Commands | `$ command` inside ` ```sh ` fence |
| Output | Verbatim after command (no `$` prefix) |
| Exit code | `# exit 0` when stdout empty |
| Annotations | `> **Annotation:**` blockquotes between fences |
| Checks | `> **Check:** ✓` or `✗` |
| Verdict | `**PASS**` or `**FAIL**` in header and Summary |
| Host install | Record method, targets, dry-run/plan, apply (see topic `host`) |

## Quick start (agent)

### Sandbox mode (default)

```sh
SLUG=my-feature
TRANSCRIPT="${SANDBOX_TRANSCRIPTS}/$(date +%Y-%m-%dT%H%M%S)-${SLUG}.md"
VERIFY_SKILL_ROOT=".agents/skills/verify-on-behalf-of-user"
source "${VERIFY_SKILL_ROOT}/scripts/enter-sandbox.sh"

# Run workflow phases; write $TRANSCRIPT with Mode: sandbox
# Then: cat "$TRANSCRIPT" into the user-facing reply (full body)
echo "Transcript: $TRANSCRIPT"
```

### Host mode (opt-in only)

```sh
# Do NOT source enter-sandbox.sh
# Resolve change-scoped targets; wrk --reinstall-local --dry-run then apply
#   or script/go install fallbacks — see topic host
SLUG=my-feature
TRANSCRIPT="${HOME}/.sandbox/transcripts/$(date +%Y-%m-%dT%H%M%S)-${SLUG}.md"
mkdir -p "$(dirname "$TRANSCRIPT")"
# write $TRANSCRIPT with Mode: host + install plan/apply
echo "Transcript: $TRANSCRIPT"
```
