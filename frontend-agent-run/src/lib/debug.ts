const DEBUG_PREFIX = '[agent-run-debug]'

export function debugLog(label: string, data?: unknown) {
  if (data === undefined) {
    console.info(DEBUG_PREFIX, label)
    return
  }
  console.info(DEBUG_PREFIX, label, data)
}

export const TERMINAL_DISCOVERY_POLL_MS = 3000
