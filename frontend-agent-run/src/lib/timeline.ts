import type { AgentEvent } from '../api/client'
import { compactProgressTimeline } from '../progressTimeline'

export function normalizeUserPromptText(text: string): string {
  return text.trim().replace(/\s+/g, ' ')
}

export function shouldSkipDuplicateUserEvent(prev: AgentEvent[], ev: AgentEvent): boolean {
  if (ev.type !== 'message' || ev.role !== 'user') {
    return false
  }
  const text = normalizeUserPromptText(ev.text ?? '')
  if (!text) {
    return false
  }
  return prev.some(
    (existing) =>
      existing.type === 'message' &&
      existing.role === 'user' &&
      normalizeUserPromptText(existing.text ?? '') === text,
  )
}

export function dedupeUserMessagesByText(events: AgentEvent[]): AgentEvent[] {
  const seen = new Set<string>()
  return events.filter((ev) => {
    if (ev.type !== 'message' || ev.role !== 'user') {
      return true
    }
    const text = normalizeUserPromptText(ev.text ?? '')
    if (!text) {
      return true
    }
    if (seen.has(text)) {
      return false
    }
    seen.add(text)
    return true
  })
}

export function appendTimelineEvent(prev: AgentEvent[], ev: AgentEvent): AgentEvent[] {
  if (shouldSkipDuplicateUserEvent(prev, ev)) {
    return prev
  }
  return dedupeUserMessagesByText([...prev, ev])
}

/**
 * Merge server events with client timeline.
 * When keepOptimisticUsers is true, user messages present only on the client
 * (composer optimistic bubbles) are appended after server events — so the
 * server list must already include prior user history. Callers must not strip
 * historical user events before merge (only the in-flight pending prompt).
 */
export function mergeSessionEvents(
  prev: AgentEvent[],
  serverEvents: AgentEvent[],
  keepOptimisticUsers: boolean,
): AgentEvent[] {
  const server = dedupeUserMessagesByText(serverEvents)
  if (!keepOptimisticUsers) {
    return server
  }
  const merged = [...server]
  for (const ev of prev) {
    if (ev.type === 'message' && ev.role === 'user') {
      if (!shouldSkipDuplicateUserEvent(merged, ev)) {
        merged.push(ev)
      }
    }
  }
  return dedupeUserMessagesByText(merged)
}

/** True when an assistant message appears before the first user message (live reorder bug). */
export function hasAssistantBeforeFirstUser(events: AgentEvent[]): boolean {
  let sawUser = false
  for (const ev of events) {
    if (ev.type !== 'message') continue
    if (ev.role === 'user' && (ev.text?.trim() ?? '')) {
      sawUser = true
      return false
    }
    if (ev.role === 'assistant' && (ev.text?.trim() ?? '') && !sawUser) {
      return true
    }
  }
  return false
}

export function summarizeEvents(events: AgentEvent[]) {
  const userMessages = events
    .filter((ev) => ev.type === 'message' && ev.role === 'user')
    .map((ev) => ev.text ?? '')
  const assistantMessages = events.filter((ev) => ev.type === 'message' && ev.role === 'assistant').length
  return {
    total: events.length,
    userCount: userMessages.length,
    assistantCount: assistantMessages,
    users: userMessages,
  }
}

export function formatMessageTimestamp(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) return ''
  return new Date(ms).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function shouldMergeAssistantStream(prev: AgentEvent, next: AgentEvent): boolean {
  if (prev.id && next.id && prev.id === next.id) {
    return true
  }
  const prevText = (prev.text ?? '').trim()
  const nextText = (next.text ?? '').trim()
  if (!prevText || !nextText) {
    return false
  }
  return nextText.startsWith(prevText)
}

export function coalesceAssistantStreaming(events: AgentEvent[]): AgentEvent[] {
  const out: AgentEvent[] = []
  for (const ev of events) {
    const last = out[out.length - 1]
    if (
      last?.type === 'message' &&
      last.role === 'assistant' &&
      ev.type === 'message' &&
      ev.role === 'assistant' &&
      shouldMergeAssistantStream(last, ev)
    ) {
      out[out.length - 1] = {
        ...ev,
        text: ev.text ?? last.text,
        id: ev.id ?? last.id,
      }
      continue
    }
    out.push(ev)
  }
  return out
}

export function isTimelineEvent(ev: AgentEvent): boolean {
  switch (ev.type) {
    case 'message':
      if (ev.role === 'user' && !(ev.text?.trim())) {
        return false
      }
      return true
    case 'think':
    case 'tool_call':
    case 'error':
      return true
    default:
      return false
  }
}

export function buildTimeline(events: AgentEvent[]): AgentEvent[] {
  const filtered = dedupeUserMessagesByText(events.filter(isTimelineEvent))
  const out: AgentEvent[] = []
  let messageBatch: AgentEvent[] = []

  const flushMessages = () => {
    if (messageBatch.length === 0) {
      return
    }
    out.push(...coalesceAssistantStreaming(messageBatch))
    messageBatch = []
  }

  for (const ev of filtered) {
    if (ev.type === 'message') {
      messageBatch.push(ev)
      continue
    }
    flushMessages()
    out.push(ev)
  }
  flushMessages()
  return compactProgressTimeline(out)
}

export function isTTYRunnerID(runner: string | undefined): boolean {
  return runner === 'codex-tty' || runner === 'grok-tty'
}
