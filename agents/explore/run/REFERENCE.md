# Explore Agent — Usage Reference

## When to use

Fast agent specialized for exploring codebases. Use this when you need to quickly find files by patterns (eg. "src/components/**/*.tsx"), search code for keywords (eg. "API endpoints"), or answer questions about the codebase (eg. "how do API endpoints work?"). When calling this agent, specify the desired thoroughness level: "quick" for basic searches, "medium" for moderate exploration, or "very thorough" for comprehensive analysis across multiple locations and naming conventions.

## When NOT to use

- If you want to read a specific file path, use the Read or Glob tool directly
- If you are searching for a specific class definition like "class Foo", use the Grep tool directly
- If you are searching within a specific file or set of 2-3 files, use the Read tool directly
- If no available agent is a good fit for the task, use other tools directly

## Usage notes

- Launch multiple agents concurrently whenever possible, to maximize performance
- Once you have delegated work to an agent, do not duplicate that work yourself
- The agent's outputs should generally be trusted
- Clearly tell the agent whether you expect it to write code or just to do research
- The explore agent is read-only: it cannot create files or modify system state
