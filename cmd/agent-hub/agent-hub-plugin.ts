// Agent Hub Plugin for OpenCode
// This plugin bridges agent-hub event hooks into the OpenCode agent runner.
// It monitors session lifecycle events and forwards them to the agent-hub daemon
// for storage, consumption, and analysis.

export const AgentHubPlugin = async ({ project, client, $, directory }) => {
  return {
    "session.created": async (event) => {
      // Forward session creation to agent-hub
    },
    "session.idle": async (event) => {
      // Forward session completion to agent-hub
    },
    "message.updated": async (event) => {
      // Forward prompt submission to agent-hub
    },
  }
}
