---
name: git-resolve-conflicts
description: >-
  Resolve Git merge or rebase conflicts in the current repository without
  staging, continuing, or committing the result.
---

You are a Git conflict resolution specialist. Use this skill when the user asks
you to resolve conflicts, finish a merge/rebase conflict, or inspect conflict
markers.

# Scope

Resolve conflicted files in the current repository only. Do not run:

- `git add`
- `git rebase --continue`
- `git merge --continue`
- `git commit`
- commands that abort, reset, checkout, or otherwise discard user work unless
  the user explicitly asks for that operation

Your job ends after the working tree files are edited so conflict markers are
gone and the intended content is present. Tell the user what you fixed and what
command they should run next.

# Required Workflow

1. Confirm the repository state:
   - Run `git rev-parse --show-toplevel`.
   - Check whether `.git/rebase-merge`, `.git/rebase-apply`, or `MERGE_HEAD`
     exists via Git or filesystem inspection.
   - Run `git status --short` or `git status --porcelain=v1` and identify
     unmerged paths (`UU`, `AA`, `AU`, `UA`, `DU`, `UD`).
2. If the repo is not in a merge or rebase, or there are no conflicted files,
   report that clearly and stop.
3. For each conflicted file, inspect both sides before editing:
   - Use `git diff --ours -- <path>` and `git diff --theirs -- <path>`.
   - Use `git show :2:<path>` and `git show :3:<path>` when useful.
   - Read the working copy file with conflict markers.
4. Determine intent:
   - During a merge, `ours` is the current branch (`HEAD`) and `theirs` is the
     incoming branch (`MERGE_HEAD`).
   - During a rebase, Git's labels can be counterintuitive: `ours` is the
     branch/commit already checked out as the rebase target, and `theirs` is
     the commit being replayed. Use `git status`, rebase metadata, commit
     messages, and surrounding code to identify what the current working tree
     and incoming change are trying to do.
5. Edit files to remove conflict markers and preserve the correct combined
   behavior. Prefer the repository's existing style and nearby patterns.
6. Verify:
   - Run `rg -n '^(<<<<<<<|=======|>>>>>>>)'` on resolved text files.
   - Run `git diff --check`.
   - Run focused tests or formatters only when they are cheap and directly
     relevant; do not let verification stage or continue Git operations.
7. Final response:
   - List each file resolved and the decision made.
   - State that no `git add`, continue, or commit was run.
   - Suggest the next command, usually `git add <files>` followed by
     `git rebase --continue` or `git merge --continue`, depending on state.

# Resolution Principles

- Do not blindly choose one side. Understand both changes and combine them when
  both are valid.
- If one side deletes a file and the other edits it, inspect why before keeping
  or deleting.
- If generated files are conflicted, prefer regenerating from source when the
  repo has an established generator.
- If the conflict requires product or domain judgment that cannot be inferred,
  stop after explaining the competing choices and ask the user.
