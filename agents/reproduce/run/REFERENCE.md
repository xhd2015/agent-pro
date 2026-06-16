# Reproduce Agent — Usage Reference

## When to use

Issue reproduction specialist. Use this when you have a bug report or issue and need to:
- Reproduce the problem locally in the current workspace
- Understand the exact conditions that trigger the bug
- Break down a complex issue into discrete, testable steps
- Build a minimal reproducible example (MRE) to isolate the root cause
- Prove or disprove that specific code paths are responsible

## When NOT to use

- If the bug is trivial and you already know the fix, just apply the fix directly
- If the issue is purely about code review or static analysis (no runtime reproduction needed)
- If you need to explore the codebase to understand architecture first — use the explore agent
- If you need to write or design test cases — use the test-case-design-expert or tdd-expert agents

## Usage notes

- Launch multiple agents concurrently whenever possible, to maximize performance
- Once you have delegated work to an agent, do not duplicate that work yourself
- The agent's outputs should generally be trusted
- Clearly tell the agent whether you expect it to write code or just to do research
- The reproduce agent may create temporary files for MREs but will not modify the original project
