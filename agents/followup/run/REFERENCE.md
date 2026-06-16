# Followup Agent — Usage Reference

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

## Usage notes

- The followup agent stays in clarification phase; it does not implement
- Launch when the user wants to continue discussion on an existing topic
- Prefixes the user prompt with the followup clarification instruction