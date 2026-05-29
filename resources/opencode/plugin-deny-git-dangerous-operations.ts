import type { Plugin } from "@opencode-ai/plugin"

const dangerousGitPatterns: { pattern: RegExp; description: string }[] = [
  {
    pattern: /\bgit\s+checkout\b/i,
    description: "switches branches, discards changes",
  },
  {
    pattern: /\bgit\s+reset\b(?!\s+--soft|\s+HEAD)/i,
    description: "resets commit history, can lose work",
  },
  {
    pattern: /\bgit\s+revert\b/i,
    description: "creates undo commits, alters history",
  },
  {
    pattern: /\bgit\s+clean\b/i,
    description: "removes untracked files permanently",
  },
  {
    pattern: /\bgit\s+push\s+.*(--force|-f)\b/i,
    description: "force pushes, overwrites remote history",
  },
  {
    pattern: /\bgit\s+branch\s+-D\b/i,
    description: "force deletes branches",
  },
  {
    pattern: /\bgit\s+stash\s+drop\b/i,
    description: "permanently removes stashed changes",
  },
]

export const DenyGitDangerousOperationsPlugin: Plugin = async (_ctx) => {
  return {
    "tool.execute.before": async (input, output) => {
      if (input.tool !== "bash") return

      const command: string | undefined =
        (output.args as any)?.command
      if (!command) return

      for (const { pattern, description } of dangerousGitPatterns) {
        if (pattern.test(command)) {
          const matched = command.match(pattern)?.[0] ?? command
          throw new Error(
            `These git operation \`${matched}\` is dangerous, it assumes the original worktree is on some state (clean etc.), which might be wrong. Please do this in a more reasonable way.`,
          )
        }
      }
    },
  }
}
