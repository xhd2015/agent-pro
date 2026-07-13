/**
 * Map finger pan deltas to xterm scrollLines amounts.
 *
 * xterm: scrollLines(n) scrolls the display down by n lines; negative scrolls up
 * (reveals older scrollback).
 *
 * Touch convention (matches content-following pan):
 * - finger moves down (positive clientY delta) → negative scrollLines (older)
 * - finger moves up (negative clientY delta) → positive scrollLines (newer)
 */

export type TouchScrollAccum = {
  lastY: number
  residualPx: number
  active: boolean
}

export function createTouchScrollAccum(): TouchScrollAccum {
  return { lastY: 0, residualPx: 0, active: false }
}

/**
 * Convert a single touchmove finger delta into integer scroll lines.
 *
 * @param fingerDyPx - clientY_new - clientY_old (positive = finger moved down the screen)
 * @param cellHeightPx - pixel height of one terminal row
 * @param residualPx - leftover sub-line pixels from prior moves
 */
export function touchDeltaToScrollLines(
  fingerDyPx: number,
  cellHeightPx: number,
  residualPx = 0,
): { lines: number; residualPx: number } {
  if (!(cellHeightPx > 0) || !Number.isFinite(cellHeightPx)) {
    return { lines: 0, residualPx }
  }
  if (!Number.isFinite(fingerDyPx)) {
    return { lines: 0, residualPx }
  }
  const total = residualPx + fingerDyPx
  // Units of cell height in the finger direction; invert for xterm scrollLines.
  const units = Math.trunc(total / cellHeightPx)
  const lines = -units
  const nextResidual = total - units * cellHeightPx
  return { lines, residualPx: nextResidual }
}

/** Prefer measured row height from the open terminal surface. */
export function measureTerminalCellHeight(
  surface: HTMLElement,
  rows: number,
  fontSize: number,
): number {
  if (rows > 0) {
    const screen =
      surface.querySelector<HTMLElement>('.xterm-screen') ||
      surface.querySelector<HTMLElement>('.xterm-rows')
    if (screen) {
      const h = screen.clientHeight || screen.getBoundingClientRect().height
      if (h > 0) {
        return h / rows
      }
    }
    const surfaceH = surface.clientHeight || surface.getBoundingClientRect().height
    if (surfaceH > 0) {
      return surfaceH / rows
    }
  }
  const fs = fontSize > 0 ? fontSize : 13
  return fs * 1.2
}
