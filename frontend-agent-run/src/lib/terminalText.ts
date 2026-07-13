export function isTerminalControlMessage(data: string): boolean {
  try {
    const parsed = JSON.parse(data) as { type?: unknown }
    return parsed?.type === 'session_id'
  } catch {
    return false
  }
}

export function sanitizeTerminalTranscript(data: string): string {
  return data
    .replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '')
    .replace(/\x1b\][^\x07]*(?:\x07|\x1b\\)/g, '')
    .replace(/[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]/g, '')
}
