import { describe, expect, it } from 'vitest'
import {
  measureTerminalCellHeight,
  touchDeltaToScrollLines,
} from './terminalTouchScroll'

describe('touchDeltaToScrollLines', () => {
  it('finger down maps to negative scrollLines (older history)', () => {
    const r = touchDeltaToScrollLines(30, 15, 0)
    expect(r.lines).toBe(-2)
    expect(r.residualPx).toBe(0)
  })

  it('finger up maps to positive scrollLines (newer)', () => {
    const r = touchDeltaToScrollLines(-30, 15, 0)
    expect(r.lines).toBe(2)
    expect(r.residualPx).toBe(0)
  })

  it('accumulates residual across small moves', () => {
    let residual = 0
    let totalLines = 0
    for (let i = 0; i < 5; i++) {
      const r = touchDeltaToScrollLines(4, 10, residual)
      totalLines += r.lines
      residual = r.residualPx
    }
    // 5 * 4 = 20px → 2 lines of height 10, inverted → -2
    expect(totalLines).toBe(-2)
    expect(residual).toBe(0)
  })

  it('returns zero lines when cell height invalid', () => {
    expect(touchDeltaToScrollLines(40, 0, 0)).toEqual({ lines: 0, residualPx: 0 })
    expect(touchDeltaToScrollLines(40, -5, 3)).toEqual({ lines: 0, residualPx: 3 })
  })

  it('keeps fractional residual under one cell', () => {
    const r = touchDeltaToScrollLines(12, 10, 0)
    expect(r.lines).toBe(-1)
    expect(r.residualPx).toBe(2)
  })
})

describe('measureTerminalCellHeight', () => {
  it('falls back to fontSize * 1.2 when rows are zero', () => {
    const el = {
      querySelector: () => null,
      clientHeight: 0,
      getBoundingClientRect: () => ({ height: 0 }),
    } as unknown as HTMLElement
    expect(measureTerminalCellHeight(el, 0, 13)).toBeCloseTo(13 * 1.2)
  })
})
