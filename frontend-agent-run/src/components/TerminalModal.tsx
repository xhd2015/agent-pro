import { useEffect, useRef, useState } from 'react'
import { FitAddon } from '@xterm/addon-fit'
import { Terminal } from '@xterm/xterm'
import '@xterm/xterm/css/xterm.css'
import { openTerminalWebSocket } from '../api/client'
import {
  isTerminalControlMessage,
  sanitizeTerminalTranscript,
} from '../lib/terminalText'
import {
  createTouchScrollAccum,
  measureTerminalCellHeight,
  touchDeltaToScrollLines,
} from '../lib/terminalTouchScroll'
import './TerminalModal.css'

export function TerminalModal({
  runner,
  sessionId,
  onClose,
}: {
  runner: string
  sessionId: string
  onClose: () => void
}) {
  const [statusText, setStatusText] = useState('terminal')
  const [terminalTranscript, setTerminalTranscript] = useState('')
  const terminalTitle = runner === 'codex-tty' ? 'Codex TTY' : runner === 'grok-tty' ? 'Grok TTY' : 'TTY'
  const showTranscriptProbe =
    terminalTranscript.includes('CODEX_TTY_BANNER') || terminalTranscript.includes('Codex')
  const wsRef = useRef<WebSocket | null>(null)
  const surfaceRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const surface = surfaceRef.current
    if (!surface) return

    const term = new Terminal({
      cursorBlink: true,
      convertEol: true,
      fontFamily: 'Menlo, Consolas, "Liberation Mono", monospace',
      fontSize: 13,
      theme: {
        background: '#05070a',
        foreground: '#e8eaed',
        cursor: '#8ab4f8',
      },
    })
    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(surface)
    const focusTerminal = () => {
      term.focus()
      surface.querySelector<HTMLElement>('.xterm-helper-textarea')?.focus()
    }
    const writeTerminal = (data: string | Uint8Array) => {
      term.write(data)
      focusTerminal()
    }

    const ws = openTerminalWebSocket(sessionId)
    wsRef.current = ws
    ws.binaryType = 'arraybuffer'
    const appendTranscript = (data: string | Uint8Array) => {
      const text = typeof data === 'string' ? data : new TextDecoder().decode(data)
      const sanitized = sanitizeTerminalTranscript(text)
      if (!sanitized) return
      setTerminalTranscript((prev) => (prev + sanitized).slice(-8000))
    }
    ws.onmessage = (event) => {
      if (typeof event.data === 'string') {
        if (isTerminalControlMessage(event.data)) {
          return
        }
        appendTranscript(event.data)
        writeTerminal(event.data)
        return
      }
      if (event.data instanceof ArrayBuffer) {
        const data = new Uint8Array(event.data)
        appendTranscript(data)
        writeTerminal(data)
      }
    }
    const sendResize = () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
      }
    }
    const fitAndResize = () => {
      fitAddon.fit()
      sendResize()
    }
    ws.onopen = fitAndResize
    ws.onerror = () => {
      setStatusText('terminal unavailable')
    }

    let inputBuffer = ''
    let inputFlushTimer = 0
    let inputVersion = 0
    const flushInput = () => {
      if (inputFlushTimer) {
        window.clearTimeout(inputFlushTimer)
        inputFlushTimer = 0
      }
      if (!inputBuffer || ws.readyState !== WebSocket.OPEN) {
        return
      }
      ws.send(new TextEncoder().encode(inputBuffer))
      inputBuffer = ''
    }
    const dataDisposable = term.onData((data) => {
      inputVersion++
      inputBuffer += data
      if (data.includes('\r') || data.includes('\n')) {
        flushInput()
        return
      }
      if (inputFlushTimer) {
        window.clearTimeout(inputFlushTimer)
      }
      inputFlushTimer = window.setTimeout(flushInput, 25)
    })
    const keydownFallback = (event: KeyboardEvent) => {
      const active = document.activeElement as HTMLElement | null
      if (
        ((active?.tagName === 'INPUT' || active?.tagName === 'TEXTAREA') &&
          !active.classList.contains('xterm-helper-textarea')) ||
        active?.isContentEditable ||
        ws.readyState !== WebSocket.OPEN
      ) {
        return
      }
      let data = ''
      if (event.key === 'Enter') {
        data = '\r'
      } else if (event.key.length === 1 && !event.metaKey && !event.ctrlKey && !event.altKey) {
        data = event.key
      }
      if (!data) return
      event.preventDefault()
      const before = inputVersion
      window.setTimeout(() => {
        if (inputVersion !== before || ws.readyState !== WebSocket.OPEN) {
          return
        }
        inputBuffer += data
        if (data === '\r') {
          flushInput()
          return
        }
        if (inputFlushTimer) {
          window.clearTimeout(inputFlushTimer)
        }
        inputFlushTimer = window.setTimeout(flushInput, 25)
      }, 0)
    }
    window.addEventListener('keydown', keydownFallback)

    // DIY mobile/touch pan → xterm scrollLines (xterm v6 is wheel-centric).
    const touchAccum = createTouchScrollAccum()
    const onTouchStart = (event: TouchEvent) => {
      if (event.touches.length !== 1) {
        touchAccum.active = false
        return
      }
      touchAccum.active = true
      touchAccum.lastY = event.touches[0].clientY
      touchAccum.residualPx = 0
    }
    const onTouchMove = (event: TouchEvent) => {
      if (!touchAccum.active || event.touches.length !== 1) {
        return
      }
      const y = event.touches[0].clientY
      const fingerDy = y - touchAccum.lastY
      touchAccum.lastY = y
      const cellH = measureTerminalCellHeight(
        surface,
        term.rows,
        typeof term.options.fontSize === 'number' ? term.options.fontSize : 13,
      )
      const { lines, residualPx } = touchDeltaToScrollLines(
        fingerDy,
        cellH,
        touchAccum.residualPx,
      )
      touchAccum.residualPx = residualPx
      if (lines !== 0) {
        term.scrollLines(lines)
        // Prevent page/modal rubber-band when we own the pan.
        if (event.cancelable) {
          event.preventDefault()
        }
      }
    }
    const onTouchEnd = () => {
      touchAccum.active = false
      touchAccum.residualPx = 0
    }
    surface.addEventListener('touchstart', onTouchStart, { passive: true })
    surface.addEventListener('touchmove', onTouchMove, { passive: false })
    surface.addEventListener('touchend', onTouchEnd, { passive: true })
    surface.addEventListener('touchcancel', onTouchEnd, { passive: true })

    const resizeObserver = new ResizeObserver(fitAndResize)
    resizeObserver.observe(surface)
    window.setTimeout(() => {
      fitAndResize()
      focusTerminal()
    }, 0)
    window.setTimeout(focusTerminal, 50)
    window.setTimeout(focusTerminal, 250)

    return () => {
      resizeObserver.disconnect()
      dataDisposable.dispose()
      window.removeEventListener('keydown', keydownFallback)
      surface.removeEventListener('touchstart', onTouchStart)
      surface.removeEventListener('touchmove', onTouchMove)
      surface.removeEventListener('touchend', onTouchEnd)
      surface.removeEventListener('touchcancel', onTouchEnd)
      if (inputFlushTimer) window.clearTimeout(inputFlushTimer)
      ws.close()
      wsRef.current = null
      term.dispose()
    }
  }, [runner, sessionId])

  return (
    <div className="terminal-modal-backdrop" role="dialog" aria-modal="true" aria-label="Terminal">
      <div className="terminal-modal">
        <div className="terminal-modal-header">
          <div className="terminal-title">{terminalTitle}</div>
          <button type="button" className="terminal-close" onClick={onClose} aria-label="Close terminal">
            Close
          </button>
        </div>
        <div
          className="terminal-surface"
          data-testid="terminal-surface"
          ref={surfaceRef}
          onClick={() => surfaceRef.current?.querySelector<HTMLElement>('.xterm-helper-textarea')?.focus()}
        />
        {showTranscriptProbe ? (
          <div className="terminal-transcript-probe" aria-hidden="true">{terminalTranscript}</div>
        ) : null}
        {statusText !== 'terminal' ? (
          <div className="terminal-status" role="status">{statusText}</div>
        ) : null}
      </div>
    </div>
  )
}
