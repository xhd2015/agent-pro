# Brainstorm Skill — Usage Reference

## When NOT to use

- Trivial task with an approach already clear from the codebase
- User explicitly asked to skip planning and go straight to implementation
- Purely reading or searching the codebase — use the explore agent
- Reproducing a bug — use the reproduce agent
- Light clarification on an existing thread — use followup

## Usage notes

- Skill-only: load via `agent-pro skill brainstorm --show` or install to `.agents/skills/`
- The parent agent follows the embedded prompt inline — no standalone binary
- Verify incrementally: the touched file or package, never the full suite — see
  `SKILL.md` §Verification
- For an HTML mockup, present the first renderable draft and judge one screenshot;
  no geometry measurement or pixel-by-pixel rounds — see `SKILL.md` §UI mockups
- Every reply ends with a context-relevant decisions section: open items each carry a
  most sound default; confirmed work shows **Locked decisions** instead
- **"go ahead"** accepts the plan plus all defaults / locked decisions unless the user
  overrode specific items — then add implementation tasks to the todo list
- For Go projects, always consider tests (doctests or unit tests) to verify correctness
