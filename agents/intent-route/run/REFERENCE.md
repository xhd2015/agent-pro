# Intent-Route Agent — Usage Reference

## When to use

Intent router that classifies user input and delegates to the appropriate
specialist agent or skill. Use this as the first step for any user request to:
- Determine whether it's a New Feature, Flash Idea, Issue/Bug, or Question
- Route New Features to the doc-style-test-based-tdd workflow
- Route Flash Ideas to the brainstorm skill for elaboration
- Route Issues/Bugs to the reproduce skill for diagnosis
- Answer questions directly without unnecessary delegation
- Clarify ambiguous requests before proceeding

## When NOT to use

- If you already know exactly which agent or skill to invoke, call it directly
- If the user explicitly names the agent or skill they want (e.g. "use
  brainstorm to expand this idea"), honor their choice
- If the request is clearly a direct question with no ambiguity, answer it
  yourself

## Usage notes

- The intent-route agent itself does not solve problems — it routes them
- Launch as the first step in a workflow to ensure the right specialist is
  invoked
- The agent outputs the category and guideline command, then proceeds
