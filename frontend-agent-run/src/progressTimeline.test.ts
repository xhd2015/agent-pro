import { describe, expect, it } from 'vitest'
import type { AgentEvent } from './api/client'
import { compactProgressTimeline, progressCardLabel } from './progressTimeline'

function tool(id: string, output?: string, ts = 0): AgentEvent {
  return {
    type: 'tool_call',
    timestamp: ts,
    text: 'Shell',
    tool: 'tool',
    tool_call_id: id,
    ...(output ? { output } : {}),
  }
}

function think(text: string, ts = 0): AgentEvent {
  return { type: 'think', timestamp: ts, text }
}

function msg(role: 'user' | 'assistant', text: string, ts = 0): AgentEvent {
  return { type: 'message', role, text, timestamp: ts }
}

function progressEvents(events: AgentEvent[]): AgentEvent[] {
  return events.filter((ev) => ev.type !== 'message')
}

function toolOutputs(events: AgentEvent[]): string[] {
  return events.filter((ev) => ev.type === 'tool_call').map((ev) => ev.output ?? '')
}

describe('compactProgressTimeline', () => {
  it('merges consecutive think events', () => {
    const input = [think('first', 1), think('second', 2)]
    const out = compactProgressTimeline(input)
    expect(out).toHaveLength(1)
    expect(out[0].text).toBe('second')
  })

  it('keeps A before B when A is updated after B starts (multi-tool seed)', () => {
    const input = [
      msg('user', 'run two tools'),
      tool('call-order-alpha', undefined, 1000),
      tool('call-order-beta', undefined, 2000),
      tool('call-order-alpha', 'alpha done', 3000),
      msg('assistant', 'Both tools finished'),
    ]
    const out = progressEvents(compactProgressTimeline(input))
    expect(out).toHaveLength(2)
    expect(out.every((ev) => progressCardLabel(ev) === 'Tool')).toBe(true)
    expect(toolOutputs(out)[0]).toContain('alpha done')
    expect(toolOutputs(out)[1]).toBe('')
  })

  it('moves merged same-id tool card to end when only think intervenes (compaction seed)', () => {
    const toolID = 'call-compact-demo'
    const input = [
      msg('user', 'run tools'),
      think('First think pass', 1000),
      think('Second think pass should replace first', 2000),
      tool(toolID, undefined, 3000),
      tool(toolID, undefined, 4000),
      think('Think between duplicate tool updates', 4500),
      tool(toolID, 'line-output\n'.repeat(2), 5000),
      msg('assistant', 'Done'),
    ]
    const out = progressEvents(compactProgressTimeline(input))
    expect(out.length).toBeGreaterThanOrEqual(2)
    expect(out.length).toBeLessThanOrEqual(4)
    expect(progressCardLabel(out[out.length - 1])).toBe('Tool')
    const labels = out.map((ev) => progressCardLabel(ev))
    const thinkIdx = labels.findIndex((l) => l === 'Thinking')
    const toolIdx = labels.findIndex((l) => l === 'Tool')
    expect(thinkIdx).toBeGreaterThanOrEqual(0)
    expect(toolIdx).toBeGreaterThanOrEqual(0)
    expect(thinkIdx).toBeLessThan(toolIdx)
  })

  it('dedupes consecutive same-text tool_call rows only for matching tool_call_id', () => {
    const input = [
      tool('same-id', undefined, 1),
      tool('same-id', undefined, 2),
    ]
    expect(compactProgressTimeline(input)).toHaveLength(1)
  })

  it('does not dedupe consecutive same-text rows for different tool_call_id', () => {
    const input = [
      tool('alpha', undefined, 1),
      tool('beta', undefined, 2),
    ]
    expect(compactProgressTimeline(input)).toHaveLength(2)
  })

  it('dedupes consecutive same-text rows when both lack tool_call_id', () => {
    const row: AgentEvent = { type: 'tool_call', text: 'Shell', tool: 'tool', timestamp: 1 }
    const input = [row, { ...row, timestamp: 2 }]
    expect(compactProgressTimeline(input)).toHaveLength(1)
  })
})