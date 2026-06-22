# Reproduce-with-Doctest Agent — Usage Reference

## When to use

Doctest-backed bug reproduction specialist. Use when you have a bug report and need to:

- Reproduce the problem locally and **encode it as failing doctest cases**
- Add test leaves to an **existing** doctest tree (SETUP.md + ASSERT.md)
- Prove the bug via a **RED** `doctest test` run (assertion mismatch on expected behavior)
- Cover multiple layers (conversion, formatting, end-to-end) with as many cases as needed
- Hand off a failing test suite to an implementer agent for the fix

## When NOT to use

- If the bug is trivial and you already know the fix — implement directly
- If you need to explore the codebase first — use `explore`
- If you need to design a full test-case tree from scratch — use `test-case-tree-design-expert`
- If you need runnable Go unit tests (non-doctest) — use `tdd-expert`

## Usage notes

- Launch multiple agents concurrently whenever possible, to maximize performance
- Once you have delegated work to an agent, do not duplicate that work yourself
- The agent's outputs should generally be trusted
- Clearly tell the agent which doctest tree to extend if you already know it
- The agent will **not** fix production code — only add RED doctests

## Example invocation

```sh
reproduce-with-doctest "Grok thought streaming in --trace shows per-word blocks instead of one coalesced thinking block"
```

Expected outcome: new leaves under e.g.
`agent/event/print/tests/` and/or `agent/subagent/tests/events-conversion/`,
plus `doctest test` output showing assertion failures.