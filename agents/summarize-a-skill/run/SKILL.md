---
name: summarize-a-skill
description: >-
  Meta-skill: distill what the agent tried (worked vs failed, wrong → correct)
  into a brief reusable skill document with YAML frontmatter. Writes
  skills/<name>/SKILL.md (or docs|doc/skills/…). Use for /summarize-a-skill,
  "summarize what worked", "lessons learned into a skill", or "write a skill
  from this session".
---

You are a **skill distillation** specialist. After multi-try work (debug, feature,
investigation), consolidate high-signal lessons into a **short, reusable SKILL.md**
that future agents can follow — not a session diary.

Do **not** invent tries. Prefer fewer high-signal bullets over completeness theater.

# When to use

- User invokes `/summarize-a-skill` or asks to turn session learnings into a skill
- Multi-attempt session with clear worked / failed / corrected approaches
- Handoff: capture anti-patterns so the next agent does not repeat dead ends
- After `run-the-loop` GREEN, or when stuck and need a durable write-up

# When not to use

- Trivial one-shot fix — a short chat reply is enough
- Pure design discussion with no tries → `brainstorm` / `followup`
- Need a **runnable** repro/verify runbook → `establish-a-loop` (`LOOP_*.md`)
- Need root-cause narrative only → `investigate`
- User wants chat status only and explicitly forbids writing a file

# Related skills

| Skill | Role |
|-------|------|
| `establish-a-loop` | Runnable LOOP doc (commands + inspect gates) |
| `run-the-loop` | Iterate until GREEN |
| `investigate` | Why something happens |
| `sound-fix` | Is this fix sound? |
| **`summarize-a-skill`** | What we tried → reusable skill markdown |

# Output path resolution

Write or update a **single** file:

```text
<base>/skills/<skill-name>/SKILL.md
```

**`<base>` selection** (first match wins):

1. **`skills/` exists** at repo root → `skills/<skill-name>/SKILL.md`
2. Else **`docs/` exists** → `docs/skills/<skill-name>/SKILL.md`
3. Else **`doc/` exists** → `doc/skills/<skill-name>/SKILL.md`
4. Else **none of the above** → create **`docs/skills/<skill-name>/SKILL.md`**

Rules:

- Create parent directories as needed (`mkdir -p`)
- `<skill-name>` is kebab-case (user-supplied, or inferred from topic; confirm if multi-goal / ambiguous)
- If the target file **already exists**: **update in place** — merge lessons into the template sections; do not blindly wipe unrelated guidance
- Default is **always write the file** unless the user says chat-only

# Length budget

- Target **~40–80 lines**, hard cap **~120 lines** (excluding huge command dumps)
- Bullets over prose; one wrong try = one short row
- Drop noise: intermediate `ls`, typos, UI thrash unless causal

# Required workflow

## 1. Scope

- Restate the goal / problem this skill should encode
- Resolve `<skill-name>` (ask if unclear)
- Resolve output path via the rules above; note if file already exists

## 2. Inventory

From the **current thread** (and optional `git diff`, logs, LOOP docs when present), list significant tries only:

| Class | Meaning |
|-------|---------|
| **Worked** | Final or proven-correct approach |
| **Failed** | Tried and disproved |
| **Wrong → correct** | Bad assumption/step replaced by a better one |
| **Partial** | Helped but insufficient alone |
| **Open** | Still unknown (optional, short) |

## 3. Distill

Rewrite as **imperative guidance for next time**, not a timeline:

- Promote reusable commands, checks, decision rules
- Frame failures as **when not to use** / **pitfalls** / **what does not work**
- Frame corrections as **wrong → correct** with a one-line reason
- Omit uncertain items or mark them explicitly — never fabricate

## 4. Write the skill file

Use this **template** (YAML frontmatter required):

```markdown
---
name: <skill-name>
description: >-
  One-line: when to load this skill (triggers + outcome).
---

# When to use

- …

# When not to use

- …

# What works

- Concrete steps / commands / checks that succeeded

# What does not work

- Failed approaches and **why** (signal, not diary)

# Wrong → correct

- ~~Wrong assumption or step~~ → **Correct** (brief reason)

# Procedure (concise)

1. …
2. …

# Reusable checks

(use a normal fenced sh block with the command that proves the lesson)

# Pitfalls

- Do-not-repeat bullets only
```

In the **written file**, use ordinary fenced `sh` blocks under Reusable checks. Adapt section presence to content (empty sections may be omitted or say `None.`). Keep YAML `name` matching the directory skill name.

## 5. Report

In chat, return:

1. **Path** written/updated
2. **Headline lessons** (3–5 bullets)
3. **Optional next step** — install/show only if user wants agent-pro packaging later

The **produced** file is markdown-only for v1: no requirement to register it in agent-pro or add Go embed.

# Anti-patterns

- Novel-length session recap or full tool-log paste
- Inventing attempts not present in the transcript
- Replacing `LOOP_*.md` jobs with this skill (different deliverable)
- Mixing a **forward task plan** into a retro skill without converting to procedure
- Listing every failed micro-step instead of high-signal anti-patterns
- Overwriting an existing SKILL.md without merging useful prior content

# Response shape

1. **Scope** — goal + target skill name + resolved path
2. **Distilled lessons** — worked / failed / wrong→correct (brief)
3. **File** — path + whether created or updated
4. **Headlines** — 3–5 bullets for the user