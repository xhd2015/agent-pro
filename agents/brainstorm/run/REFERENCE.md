# Brainstorm Skill — Usage Reference

## When to use

Brainstorming specialist. Use this before implementing any feature or fix when you need to:
- Discuss the approach and trade-offs with the user before writing code
- Plan data models, storage layouts, and architecture
- Define test scenarios and expected outputs
- Plan CLI output examples when designing commands or sub-commands
- Get explicit user confirmation before proceeding to implementation

## When NOT to use

- If the task is trivial and the approach is already clear from the codebase
- If the user explicitly asked to skip planning and go straight to implementation
- If the task is purely about reading or searching the codebase — use the explore agent
- If the task is about reproducing a bug — use the reproduce agent
- Light clarification on an existing thread — use followup

## Usage notes

- Skill-only: load via `agent-pro skill brainstorm --show` or install to `.agents/skills/`
- The parent agent follows the embedded prompt inline — no standalone binary
- Once the user confirms with "go ahead", add implementation tasks to the todo list
- For Go projects, always consider tests (doctests or unit tests) to verify correctness