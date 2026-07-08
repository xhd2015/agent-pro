import { describe, expect, it } from 'vitest'
import type { SessionSummary } from './api/client'
import { sessionsPollResult } from './sessionPoll'

function session(id: string): SessionSummary {
  return { runner: 'opencode', session_id: id, status: 'running' }
}

describe('sessionsPollResult', () => {
  it('replaces list when fetch succeeds', () => {
    const next = [session('new')]
    expect(sessionsPollResult(next, [session('old')])).toEqual(next)
  })

  it('preserves current list when fetch fails', () => {
    const current = [session('keep')]
    expect(sessionsPollResult(null, current)).toBe(current)
  })

  it('accepts empty array from successful fetch', () => {
    expect(sessionsPollResult([], [session('old')])).toEqual([])
  })
})