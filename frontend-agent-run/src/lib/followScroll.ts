export const BOTTOM_THRESHOLD_PX = 80
export const TOP_THRESHOLD_PX = 80

export function distanceFromBottom(el: HTMLElement): number {
  return el.scrollHeight - el.scrollTop - el.clientHeight
}

export function distanceFromTop(el: HTMLElement): number {
  return el.scrollTop
}
