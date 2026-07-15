# Followup Skill — Usage Reference

## When to use

Clarification-phase followup. Use when the user mentions followup to:
- Continue discussing an approach without starting implementation
- Ask clarifying questions or refine requirements
- Stay in planning or clarification mode after an initial brainstorm

## When NOT to use

- If the user explicitly asked to implement or said "go ahead"
- If the task is a new request that needs routing — use intent-route first
- If the task needs codebase exploration — use explore
- If the task is about reproducing a bug — use reproduce
- If the topic needs full planning — use brainstorm

## Usage notes

- Skill-only: load via `agent-pro skill followup --show` or install to `.agents/skills/`
- The parent agent follows the embedded prompt inline — no standalone binary
- Stays in clarification phase; does not implement
- Launch when the user wants to continue discussion on an existing topic