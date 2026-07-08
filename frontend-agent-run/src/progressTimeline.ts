import type { AgentEvent } from './api/client'

export function sanitizeProgressText(text: string): string {
  return text
    .replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '')
    .replace(/[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]/g, '')
    .trim()
}

export function truncateProgressText(text: string, maxLen = 140): string {
  const clean = sanitizeProgressText(text)
  if (!clean) {
    return ''
  }
  const lines = clean.split(/\r?\n/).filter((line) => line.trim())
  if (lines.length > 2) {
    const preview = lines.slice(0, 2).join('\n')
    const suffix = ` (+${lines.length - 2} more lines)`
    const budget = maxLen - suffix.length
    if (preview.length + suffix.length <= maxLen) {
      return preview + suffix
    }
    return preview.slice(0, Math.max(budget, 24)) + '…' + suffix
  }
  if (clean.length <= maxLen) {
    return clean
  }
  return clean.slice(0, maxLen - 1) + '…'
}

function mergeToolCall(prev: AgentEvent, ev: AgentEvent): AgentEvent {
  return {
    ...ev,
    text: ev.text ?? prev.text,
    output: ev.output ?? prev.output,
    tool: ev.tool ?? prev.tool,
  }
}

function findPriorToolCallIndex(out: AgentEvent[], toolCallID: string): number {
  for (let i = out.length - 1; i >= 0; i--) {
    if (out[i].type === 'message') {
      break
    }
    if (out[i].type === 'tool_call' && out[i].tool_call_id?.trim() === toolCallID) {
      return i
    }
  }
  return -1
}

function hasInterveningDifferentTool(out: AgentEvent[], existingIdx: number, toolCallID: string): boolean {
  for (let i = existingIdx + 1; i < out.length; i++) {
    if (out[i].type !== 'tool_call') {
      continue
    }
    const otherID = out[i].tool_call_id?.trim()
    if (otherID && otherID !== toolCallID) {
      return true
    }
  }
  return false
}

function shouldSkipConsecutiveDuplicateTool(last: AgentEvent, ev: AgentEvent): boolean {
  if (last.type !== 'tool_call' || ev.type !== 'tool_call') {
    return false
  }
  if (last.output?.trim() || ev.output?.trim()) {
    return false
  }
  const lastID = last.tool_call_id?.trim()
  const evID = ev.tool_call_id?.trim()
  const sameToolID = Boolean(lastID && evID && lastID === evID)
  const bothMissingID = !lastID && !evID
  if (!sameToolID && !bothMissingID) {
    return false
  }
  return progressCardText(last) === progressCardText(ev)
}

export function compactProgressTimeline(events: AgentEvent[]): AgentEvent[] {
  const out: AgentEvent[] = []
  for (const ev of events) {
    if (ev.type === 'message') {
      out.push(ev)
      continue
    }
    const last = out[out.length - 1]
    if (ev.type === 'think') {
      if (last?.type === 'think') {
        out[out.length - 1] = ev
        continue
      }
      out.push(ev)
      continue
    }
    if (ev.type === 'tool_call') {
      const toolCallID = ev.tool_call_id?.trim()
      if (toolCallID) {
        const existingIdx = findPriorToolCallIndex(out, toolCallID)
        if (existingIdx >= 0) {
          const merged = mergeToolCall(out[existingIdx], ev)
          if (hasInterveningDifferentTool(out, existingIdx, toolCallID)) {
            out[existingIdx] = merged
          } else {
            out.splice(existingIdx, 1)
            out.push(merged)
          }
          continue
        }
      }
      if (last && shouldSkipConsecutiveDuplicateTool(last, ev)) {
        continue
      }
      out.push(ev)
      continue
    }
    out.push(ev)
  }
  return out
}

export function progressCardLabel(ev: AgentEvent): string {
  if (ev.type === 'think') {
    return 'Thinking'
  }
  if (ev.type === 'tool_call') {
    return 'Tool'
  }
  return 'Working'
}

export function progressCardText(ev: AgentEvent): string {
  if (ev.type === 'think') {
    return truncateProgressText(ev.text?.trim() || 'Thinking…')
  }
  if (ev.type === 'tool_call') {
    const parts = [ev.tool?.trim(), ev.text?.trim(), ev.output?.trim()].filter(Boolean)
    if (parts.length > 0) {
      return truncateProgressText(parts.join(': '))
    }
    return 'Tool call'
  }
  return truncateProgressText(ev.text?.trim() || 'Working…')
}