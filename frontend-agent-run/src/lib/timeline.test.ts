import { describe, expect, it } from 'vitest'
import type { AgentEvent } from '../api/client'
import {
  appendTimelineEvent,
  buildTimeline,
  coalesceAssistantStreaming,
  dedupeUserMessagesByText,
  mergeSessionEvents,
  normalizeUserPromptText,
  shouldSkipDuplicateUserEvent,
} from './timeline'

function msg(role: 'user' | 'assistant', text: string, extra: Partial<AgentEvent> = {}): AgentEvent {
  return { type: 'message', role, text, ...extra }
}

describe('timeline helpers', () => {
  it('normalizes user prompt text', () => {
    expect(normalizeUserPromptText('  hello   world  ')).toBe('hello world')
  })

  it('skips duplicate user events by normalized text', () => {
    const prev = [msg('user', 'hello world')]
    expect(shouldSkipDuplicateUserEvent(prev, msg('user', '  hello   world  '))).toBe(true)
    expect(shouldSkipDuplicateUserEvent(prev, msg('user', 'other'))).toBe(false)
    expect(shouldSkipDuplicateUserEvent(prev, msg('assistant', 'hello world'))).toBe(false)
  })

  it('dedupes user messages by text while keeping assistants', () => {
    const events = [
      msg('user', 'a'),
      msg('assistant', 'reply'),
      msg('user', 'a'),
      msg('user', 'b'),
    ]
    expect(dedupeUserMessagesByText(events)).toEqual([
      msg('user', 'a'),
      msg('assistant', 'reply'),
      msg('user', 'b'),
    ])
  })

  it('appendTimelineEvent drops duplicate user and appends new', () => {
    const prev = [msg('user', 'hi')]
    expect(appendTimelineEvent(prev, msg('user', 'hi'))).toEqual(prev)
    expect(appendTimelineEvent(prev, msg('assistant', 'yo'))).toEqual([
      msg('user', 'hi'),
      msg('assistant', 'yo'),
    ])
  })

  it('mergeSessionEvents keeps optimistic users when requested', () => {
    const prev = [msg('user', 'optimistic'), msg('assistant', 'local')]
    const server = [msg('assistant', 'server')]
    expect(mergeSessionEvents(prev, server, true)).toEqual([
      msg('assistant', 'server'),
      msg('user', 'optimistic'),
    ])
    expect(mergeSessionEvents(prev, server, false)).toEqual([msg('assistant', 'server')])
  })

  it('mergeSessionEvents follow-up: keep prior users on server + append pending optimistic', () => {
    // Server has turn-1 history; only the in-flight follow-up is client-only.
    const prev = [
      msg('user', 'run ls'),
      msg('assistant', 'files…'),
      msg('user', 'what did I say'),
    ]
    const server = [msg('user', 'run ls'), msg('assistant', 'files…')]
    expect(mergeSessionEvents(prev, server, true)).toEqual([
      msg('user', 'run ls'),
      msg('assistant', 'files…'),
      msg('user', 'what did I say'),
    ])
  })

  it('mergeSessionEvents does not put prior user after assistants when server kept them', () => {
    const prev = [
      msg('user', 'run ls'),
      msg('assistant', 'MOCK'),
      msg('user', 'what did I say'),
    ]
    // Wrong historical bug: server stripped ALL users → assistants then re-appended users.
    const serverStrippedAllUsers = [msg('assistant', 'MOCK')]
    const bad = mergeSessionEvents(prev, serverStrippedAllUsers, true)
    expect(bad[0]?.role).toBe('assistant') // documents failure mode if caller strips all

    // Correct call: only pending follow-up omitted from server.
    const serverPendingOnly = [msg('user', 'run ls'), msg('assistant', 'MOCK')]
    const good = mergeSessionEvents(prev, serverPendingOnly, true)
    expect(good.map((e) => e.role)).toEqual(['user', 'assistant', 'user'])
    expect(good.map((e) => e.text)).toEqual(['run ls', 'MOCK', 'what did I say'])
  })

  it('coalesceAssistantStreaming merges growing assistant chunks', () => {
    const events = [
      msg('assistant', 'Hel', { id: '1' }),
      msg('assistant', 'Hello', { id: '1' }),
      msg('user', 'next'),
      msg('assistant', 'Hi'),
    ]
    expect(coalesceAssistantStreaming(events)).toEqual([
      msg('assistant', 'Hello', { id: '1' }),
      msg('user', 'next'),
      msg('assistant', 'Hi'),
    ])
  })

  it('buildTimeline filters empty user messages and coalesces streams', () => {
    const events: AgentEvent[] = [
      msg('user', ''),
      msg('user', 'hi'),
      { type: 'think', text: 'thinking' },
      msg('assistant', 'He', { id: 'a' }),
      msg('assistant', 'Hello', { id: 'a' }),
      { type: 'unknown_skip' as AgentEvent['type'] },
    ]
    const timeline = buildTimeline(events)
    expect(timeline.some((ev) => ev.type === 'message' && ev.role === 'user' && !ev.text?.trim())).toBe(
      false,
    )
    const assistants = timeline.filter((ev) => ev.type === 'message' && ev.role === 'assistant')
    expect(assistants.length).toBe(1)
    expect(assistants[0]?.text).toBe('Hello')
  })
})
