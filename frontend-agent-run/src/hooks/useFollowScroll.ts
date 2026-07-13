import {
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  type RefObject,
} from 'react'
import {
  BOTTOM_THRESHOLD_PX,
  TOP_THRESHOLD_PX,
  distanceFromBottom,
  distanceFromTop,
} from '../lib/followScroll'

export type FollowScrollAnchor = 'top' | 'bottom'

export type FollowScrollOptions = {
  /** Where "latest" content lives. Home newest-first uses top; chat uses bottom. */
  anchor?: FollowScrollAnchor
  /**
   * On first layout, restore this scrollTop and start detached (Home remount after
   * navigating to a session and back). Ignored when ≤ threshold (at latest).
   */
  restoreScrollTop?: number
}

export function useFollowScroll<T extends HTMLElement>(
  scrollRef: RefObject<T | null>,
  contentDeps: unknown[],
  options?: FollowScrollOptions,
) {
  const anchor: FollowScrollAnchor = options?.anchor === 'top' ? 'top' : 'bottom'
  const threshold = anchor === 'top' ? TOP_THRESHOLD_PX : BOTTOM_THRESHOLD_PX
  const restoreScrollTop = options?.restoreScrollTop ?? 0
  const followModeRef = useRef<'following' | 'detached'>('following')
  const pinnedScrollTopRef = useRef<number | null>(null)
  const isProgrammaticScrollRef = useRef(false)
  const userScrollIntentRef = useRef(false)
  const lastScrollTopRef = useRef(0)
  const initialScrollDoneRef = useRef(false)
  const restoreAppliedRef = useRef(false)
  const [followMode, setFollowMode] = useState<'following' | 'detached'>('following')
  const [showJumpToLatest, setShowJumpToLatest] = useState(false)

  const distanceFromLatest = useCallback(
    (el: HTMLElement) => (anchor === 'top' ? distanceFromTop(el) : distanceFromBottom(el)),
    [anchor],
  )

  const scrollToLatest = useCallback(
    (el: HTMLElement) => {
      el.scrollTop = anchor === 'top' ? 0 : el.scrollHeight
    },
    [anchor],
  )

  const markUserScrollIntent = useCallback(() => {
    userScrollIntentRef.current = true
  }, [])

  const applyFollowMode = useCallback((mode: 'following' | 'detached') => {
    followModeRef.current = mode
    setFollowMode(mode)
    if (mode === 'following') {
      pinnedScrollTopRef.current = null
      setShowJumpToLatest(false)
    } else {
      setShowJumpToLatest(true)
    }
  }, [])

  const resetFollow = useCallback(() => {
    initialScrollDoneRef.current = false
    pinnedScrollTopRef.current = null
    applyFollowMode('following')
  }, [applyFollowMode])

  const restorePinnedScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el || pinnedScrollTopRef.current == null) return
    isProgrammaticScrollRef.current = true
    el.scrollTop = pinnedScrollTopRef.current
    lastScrollTopRef.current = el.scrollTop
  }, [scrollRef])

  const syncFollowFromScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el) return

    // Swallow the scroll event that results from our own scrollToLatest/pin restore.
    // Still update lastScrollTop so a coalesced real user scroll is measured correctly.
    if (isProgrammaticScrollRef.current) {
      isProgrammaticScrollRef.current = false
      lastScrollTopRef.current = el.scrollTop
      return
    }

    const prevScrollTop = lastScrollTopRef.current
    lastScrollTopRef.current = el.scrollTop
    const scrollingAwayFromLatest =
      anchor === 'top'
        ? el.scrollTop > prevScrollTop + 1
        : el.scrollTop < prevScrollTop - 1
    const distance = distanceFromLatest(el)
    const atLatest = distance <= threshold

    if (atLatest) {
      // Settled (or scrolled) at latest — follow. Do not restore a mid-list pin here;
      // that fought trackpad momentum and snapped the list after the user stopped.
      if (userScrollIntentRef.current) {
        userScrollIntentRef.current = false
      }
      applyFollowMode('following')
      return
    }

    if (userScrollIntentRef.current) {
      userScrollIntentRef.current = false
      pinnedScrollTopRef.current = el.scrollTop
      applyFollowMode('detached')
      return
    }

    if (followModeRef.current === 'detached' && pinnedScrollTopRef.current != null) {
      // Do NOT restore pin on every scroll-without-intent.
      // Trackpad/touch momentum and scrollbar drags fire scroll events without a
      // fresh wheel/touchstart; restoring the old pin causes "scroll to A, hang,
      // scroll to B, hang, settle on C, then snap back to B".
      // Content-driven jumps are corrected in the contentDeps layout effect instead.
      if (Math.abs(el.scrollTop - pinnedScrollTopRef.current) > 1) {
        pinnedScrollTopRef.current = el.scrollTop
      }
      return
    }

    // Following and content grew without user scroll away — stay following.
    if (followModeRef.current === 'following' && !scrollingAwayFromLatest) {
      return
    }

    pinnedScrollTopRef.current = el.scrollTop
    applyFollowMode('detached')
  }, [scrollRef, applyFollowMode, anchor, distanceFromLatest, threshold])

  const handleJumpToLatest = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    isProgrammaticScrollRef.current = true
    scrollToLatest(el)
    lastScrollTopRef.current = el.scrollTop
    initialScrollDoneRef.current = true
    applyFollowMode('following')
  }, [scrollRef, applyFollowMode, scrollToLatest])

  const pinDetachedScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    if (pinnedScrollTopRef.current == null) {
      pinnedScrollTopRef.current = el.scrollTop
    } else {
      isProgrammaticScrollRef.current = true
      el.scrollTop = pinnedScrollTopRef.current
    }
    lastScrollTopRef.current = el.scrollTop
    applyFollowMode('detached')
  }, [scrollRef, applyFollowMode])

  /** Pin whatever scrollTop is now (or optional explicit top) as detached. */
  const pinScrollAt = useCallback(
    (top?: number) => {
      const el = scrollRef.current
      if (!el) return
      const next = top ?? el.scrollTop
      isProgrammaticScrollRef.current = true
      el.scrollTop = next
      pinnedScrollTopRef.current = el.scrollTop
      lastScrollTopRef.current = el.scrollTop
      applyFollowMode('detached')
    },
    [scrollRef, applyFollowMode],
  )

  useLayoutEffect(() => {
    const el = scrollRef.current
    if (!el) {
      pinnedScrollTopRef.current = null
      followModeRef.current = 'following'
      setShowJumpToLatest(false)
      return
    }

    // Remount restore: pin detached scroll before "following" jumps to latest.
    if (!restoreAppliedRef.current && restoreScrollTop > threshold) {
      restoreAppliedRef.current = true
      isProgrammaticScrollRef.current = true
      el.scrollTop = restoreScrollTop
      lastScrollTopRef.current = el.scrollTop
      pinnedScrollTopRef.current = el.scrollTop
      followModeRef.current = 'detached'
      setFollowMode('detached')
      setShowJumpToLatest(true)
      initialScrollDoneRef.current = true
      return
    }

    if (pinnedScrollTopRef.current != null) {
      isProgrammaticScrollRef.current = true
      el.scrollTop = pinnedScrollTopRef.current
      lastScrollTopRef.current = el.scrollTop
      const distance = distanceFromLatest(el)
      setShowJumpToLatest(distance > threshold)
      return
    }

    const distance = distanceFromLatest(el)
    if (followModeRef.current === 'following') {
      isProgrammaticScrollRef.current = true
      scrollToLatest(el)
      lastScrollTopRef.current = el.scrollTop
      initialScrollDoneRef.current = true
      setShowJumpToLatest(false)
      return
    }

    setShowJumpToLatest(distance > threshold)
    // eslint-disable-next-line react-hooks/exhaustive-deps -- contentDeps drives follow scroll updates
  }, contentDeps)

  return {
    followModeRef,
    followMode,
    showJumpToLatest,
    syncFollowFromScroll,
    markUserScrollIntent,
    handleJumpToLatest,
    resetFollow,
    pinDetachedScroll,
    pinScrollAt,
    restorePinnedScroll,
  }
}
