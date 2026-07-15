---
name: summarize-a-skill
description: >-
  Meta-skill: distill durable high-signal guidance (worked vs failed, wrong →
  correct, and anything else that quickens future similar tasks) into a brief
  reusable skill document with YAML frontmatter. When the user does not
  explicitly ask for a new skill/file, prefer updating an existing skill over
  creating one. Writes skills/<name>/SKILL.md (or docs|doc/skills/…). Use for
  /summarize-a-skill, "summarize what worked", "lessons learned into a skill",
  or "write a skill from this session".
---

You are a **skill distillation** specialist. After multi-try work (debug, feature,
investigation), consolidate high-signal lessons into a **short, reusable SKILL.md**
that future agents can follow — not a session diary.

Capture **durable guidance that shortens the next similar task**: proven and failed
approaches **and** other worth-recording signal (paths, checks, decision rules, env
or tool quirks, “do this first”). Do **not** invent tries or facts. Prefer fewer
high-signal bullets over completeness theater.

# When to use

- User invokes `/summarize-a-skill` or asks to turn session learnings into a skill
- Multi-attempt session with clear worked / failed / corrected approaches
- Session produced **durable shortcuts** even without many failed tries (paths,
  checks, decision rules, env/tool quirks worth encoding)
- Handoff: capture anti-patterns so the next agent does not repeat dead ends
- After `run-the-loop` GREEN, or when stuck and need a durable write-up

# When not to use

- Trivial one-shot fix — a short chat reply is enough
- Lesson is one-off and **not** reusable for future similar tasks (chat only)
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
| **`summarize-a-skill`** | What we tried + accelerators → reusable skill markdown |

# Update vs create

When the user has **not** explicitly asked for a **new** skill or file, **prefer
updating an existing skill** over inventing a new one.

## Explicit create

Treat as create when the user:

- Says “create a new skill”, “new SKILL.md”, “don’t touch existing skills”, or
  supplies a **new** skill name for a separate home
- Points at a path that should be written as a new skill

Still **merge** if that target path already exists (do not wipe prior content).

## Default (no explicit create)

1. **Search candidates** (collect real hits only):
   - Skills already named, linked, or loaded in this thread
   - Repo skill trees (same bases as output resolution):
     `skills/`, `docs/skills/`, `doc/skills/`
   - Do **not** full-scan global installs (e.g. `~/.agents/skills`) unless the
     session already used a skill from there or the user points at one
2. **Decide** with these criteria:

| Prefer **update** | Prefer **create new** |
|-------------------|------------------------|
| Lessons fit an existing skill’s **when to use** / domain | Topic is **distinct**; would confuse or bloat an existing skill |
| Same tool, failure mode, or procedure family | Orthogonal workflow (different triggers) |
| File already exists for the resolved skill name | Large new procedure that deserves its own name |
| Small additive lessons | User later confirms they want a separate skill |

3. **Act**:
   - **One clear home** → update that file (merge into sections)
   - **Zero or many good homes** → **ask once** before writing
   - **Create** only when domain is distinct or user confirms a new skill
4. State the **decision and one-line reason** in Scope and in the final Report

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
- Resolve **update vs create** first (section above); path rules apply to the chosen skill name
- If the target file **already exists**: **update in place** — merge lessons into the template sections; do not blindly wipe unrelated guidance
- Default is **always write the file** unless the user says chat-only

# Length budget

- Target **~40–80 lines**, hard cap **~120 lines** (excluding huge command dumps)
- Bullets over prose; one wrong try or accelerator = one short row
- Drop noise: intermediate `ls`, typos, UI thrash unless causal

# Required workflow

## 1. Scope

- Restate the goal / problem this skill should encode
- Resolve **update vs create** (search + decide, or ask if ambiguous)
- Resolve `<skill-name>` (ask if unclear)
- Resolve output path via the rules above; note if file already exists

## 2. Inventory

From the **current thread** (and optional `git diff`, logs, LOOP docs when present),
list significant items only — **tries and anything else worth recording**:

| Class | Meaning |
|-------|---------|
| **Worked** | Final or proven-correct approach |
| **Failed** | Tried and disproved |
| **Wrong → correct** | Bad assumption/step replaced by a better one |
| **Partial** | Helped but insufficient alone |
| **Open** | Still unknown (optional, short) |
| **Accelerator** | Speeds the next run: commands, flags, paths, env vars, “check X first”, tool quirks, decision rules, prerequisites, good defaults |

## 3. Distill

Rewrite as **imperative guidance for next time**, not a timeline:

- Promote reusable commands, checks, decision rules, and look-here-first pointers
- Include any durable fact that **quickens future similar tasks**, not only multi-try narrative
- Frame failures as **when not to use** / **pitfalls** / **what does not work**
- Frame corrections as **wrong → correct** with a one-line reason
- Prefer items that change **next-time behavior**; drop pure narrative unless it encodes a rule
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

# Worth recording (accelerators)

- Paths, checks, decision rules, env/tool quirks that shorten the next run
  (optional: use this section only when ≥3 accelerator bullets; otherwise fold
  into What works / Procedure / Reusable checks / Pitfalls)

# Pitfalls

- Do-not-repeat bullets only
```

In the **written file**, use ordinary fenced `sh` blocks under Reusable checks.
**Accelerators:** if fewer than **3** bullets, fold into existing sections; if **≥3**,
use `# Worth recording (accelerators)`. Adapt section presence to content (empty
sections may be omitted or say `None.`). Keep YAML `name` matching the directory
skill name.

## 5. Report

In chat, return:

1. **Decision** — updated vs created, path, and **one-line reason** (candidates considered if relevant)
2. **Path** written/updated
3. **Headline lessons** (3–5 bullets) — include notable accelerators, not only tries
4. **Optional next step** — install/show only if user wants agent-pro packaging later

The **produced** file is markdown-only for v1: no requirement to register it in agent-pro or add Go embed.

# Anti-patterns

- Novel-length session recap or full tool-log paste
- Inventing attempts not present in the transcript
- Defaulting to a **new** `skills/<name>/SKILL.md` without checking whether an existing skill is the right home
- Replacing `LOOP_*.md` jobs with this skill (different deliverable)
- Mixing a **forward task plan** into a retro skill without converting to procedure
- Listing every failed micro-step instead of high-signal anti-patterns
- Overwriting an existing SKILL.md without merging useful prior content
- Dumping every path or command touched instead of high-signal accelerators

# Response shape

1. **Scope** — goal + update-vs-create decision + target skill name + resolved path
2. **Distilled lessons** — worked / failed / wrong→correct / accelerators (brief)
3. **File** — path + whether created or updated + reason
4. **Headlines** — 3–5 bullets for the user
