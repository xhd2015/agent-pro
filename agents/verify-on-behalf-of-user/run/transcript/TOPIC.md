---
name: verify-on-behalf-of-user/transcript
description: >-
  Markdown transcript format rules, template usage, and quick-start commands for
  verify-on-behalf-of-user.
---

# Transcript

The primary deliverable is a markdown file the user can review like a terminal session.

## Write steps

1. Start from `templates/transcript.md` (installed with the skill)
2. Fill header table (scope, verdict, paths)
3. Append each phase as fenced ` ```sh ` blocks with `$ ` prompts
4. Insert `> **Annotation:**` / `> **Check:** ✓|✗` between blocks
5. Save to `~/.sandbox/transcripts/<ISO8601>-<slug>.md`
6. Tell the user the transcript path and one-line verdict

## Transcript format rules

| Element | Format |
|---------|--------|
| Commands | `$ command` inside ` ```sh ` fence |
| Output | Verbatim after command (no `$` prefix) |
| Exit code | `# exit 0` when stdout empty |
| Annotations | `> **Annotation:**` blockquotes between fences |
| Checks | `> **Check:** ✓` or `✗` |
| Verdict | `**PASS**` or `**FAIL**` in header and Summary |

## Quick start (agent)

```sh
SLUG=my-feature
TRANSCRIPT="${SANDBOX_TRANSCRIPTS}/$(date +%Y-%m-%dT%H%M%S)-${SLUG}.md"
VERIFY_SKILL_ROOT=".agents/skills/verify-on-behalf-of-user"
source "${VERIFY_SKILL_ROOT}/scripts/enter-sandbox.sh"

# Run workflow phases; write $TRANSCRIPT
echo "Transcript: $TRANSCRIPT"
```