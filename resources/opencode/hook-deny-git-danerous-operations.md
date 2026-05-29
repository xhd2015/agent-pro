---
name: deny-git-dangerous-operations
description: Enforces confirmation for dangerous git operations
---

The following git operations are potentially dangerous and require user confirmation before execution:

- git checkout    — switches branches, discards changes
- git reset       — resets commit history, can lose work
- git revert      — creates undo commits, alters history
- git clean       — removes untracked files permanently
- git push --force / -f — force pushes, overwrites remote history
- git branch -D   — force deletes branches
- git stash drop  — permanently removes stashed changes

Always confirm with the user before executing any of these commands.

User request: $ARGUMENTS