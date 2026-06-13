// Agent Hub Plugin for OpenCode
// This plugin bridges agent-hub event hooks into the OpenCode agent runner.
// It monitors session lifecycle events and forwards them to the agent-hub daemon
// for storage, consumption, and analysis.

const { execSync } = require("child_process");

function notify(eventType: string, payload: any) {
  try {
    execSync("agent-hub hook notify --runner opencode --event " + eventType, {
      input: JSON.stringify(payload),
      stdio: ["pipe", "pipe", "pipe"],
    });
  } catch (e) {}
}

export const AgentHubPlugin = async ({ project, client, $, directory }) => {
  return {
    "session.created": async (event) => {
      notify("session.created", event);
    },
    "session.idle": async (event) => {
      notify("session.idle", event);
    },
    "message.updated": async (event) => {
      notify("message.updated", event);
    },
    "session.error": async (event) => {
      notify("session.error", event);
    },
  };
};
