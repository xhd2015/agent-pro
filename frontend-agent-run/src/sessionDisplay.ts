import type { SessionSummary } from './api/client'

export function parseRFC3339Ms(iso: string | undefined): number | null {
  if (!iso?.trim()) return null
  const ms = Date.parse(iso)
  return Number.isFinite(ms) ? ms : null
}

export function truncateSessionPreview(text: string, maxLen = 96): string {
  const normalized = text.replace(/\s+/g, ' ').trim()
  if (!normalized) return ''
  if (normalized.length <= maxLen) return normalized
  return `${normalized.slice(0, maxLen - 1)}…`
}

export function shortSessionId(sessionId: string): string {
  const id = sessionId.trim()
  if (id.length <= 20) return id
  return `${id.slice(0, 10)}…${id.slice(-8)}`
}

export function shortWorkspaceLabel(workspace: string): string {
  const trimmed = workspace.trim().replace(/\/+$/, '')
  if (!trimmed) return ''
  const parts = trimmed.split(/[/\\]/).filter(Boolean)
  if (parts.length === 0) return trimmed
  if (parts.length <= 2) return parts.join('/')
  return `…/${parts.slice(-2).join('/')}`
}

export function sessionRowLabel(session: SessionSummary): string {
  const prompt = truncateSessionPreview(session.initial_prompt ?? '')
  if (prompt) return prompt
  return shortSessionId(session.session_id)
}

export function formatSessionRecency(
  updatedAt: string | undefined,
  createdAt: string | undefined,
  nowMs = Date.now(),
): string {
  const ms = parseRFC3339Ms(updatedAt) ?? parseRFC3339Ms(createdAt)
  if (ms == null) return ''
  const deltaMs = Math.max(0, nowMs - ms)
  const deltaSec = Math.floor(deltaMs / 1000)
  if (deltaSec < 45) return 'just now'
  const deltaMin = Math.floor(deltaSec / 60)
  if (deltaMin < 60) return `${deltaMin}m ago`
  const deltaHr = Math.floor(deltaMin / 60)
  if (deltaHr < 24) return `${deltaHr}h ago`
  const deltaDay = Math.floor(deltaHr / 24)
  if (deltaDay === 1) return 'yesterday'
  if (deltaDay < 7) return `${deltaDay}d ago`
  return new Date(ms).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
  })
}

export function sortSessionsOldestFirst(sessions: SessionSummary[]): SessionSummary[] {
  return [...sessions].sort((a, b) => {
    const aMs = parseRFC3339Ms(a.updated_at) ?? parseRFC3339Ms(a.created_at) ?? 0
    const bMs = parseRFC3339Ms(b.updated_at) ?? parseRFC3339Ms(b.created_at) ?? 0
    if (aMs !== bMs) return aMs - bMs
    return a.session_id.localeCompare(b.session_id)
  })
}

export function statusPillClass(status: string): string {
  switch (status.trim().toLowerCase()) {
    case 'running':
      return 'status-pill status-pill-running'
    case 'error':
      return 'status-pill status-pill-error'
    case 'finished':
    case 'idle':
      return 'status-pill status-pill-idle'
    default:
      return 'status-pill'
  }
}