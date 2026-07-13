import { useCallback, useState } from 'react'
import { formatWorkspaceLabel } from '../sessionDisplay'
import './WorkspacePath.css'

export type WorkspacePathProps = {
  path: string
  /** Optional className on the root region (layout hooks from parent). */
  className?: string
  /**
   * When set, primary click opens the workspace selector (browse page)
   * instead of expanding the label. Used on Home.
   */
  onOpenSelector?: () => void
  /** Show the Copy control (default true when no onOpenSelector). */
  showCopy?: boolean
}

export function WorkspacePath({
  path,
  className,
  onOpenSelector,
  showCopy,
}: WorkspacePathProps) {
  const [expanded, setExpanded] = useState(false)
  const [copied, setCopied] = useState(false)
  const trimmed = path.trim()
  const copyVisible = showCopy ?? onOpenSelector == null

  const toggle = useCallback(() => {
    setExpanded((prev) => !prev)
  }, [])

  const copyPath = useCallback(async () => {
    const value = path.trim()
    if (!value) return
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(value)
      } else {
        // Fallback for environments without clipboard API
        const ta = document.createElement('textarea')
        ta.value = value
        ta.setAttribute('readonly', '')
        ta.style.position = 'fixed'
        ta.style.left = '-9999px'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
      }
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard may be denied; still no-op without throwing
    }
  }, [path])

  if (!trimmed) return null

  // Top-bar selector: always show full path (wrap/clamp in CSS). Session page:
  // collapsed = compact, expanded = full.
  const label = onOpenSelector
    ? formatWorkspaceLabel(trimmed, { mode: 'full' })
    : expanded
      ? formatWorkspaceLabel(trimmed, { mode: 'full' })
      : formatWorkspaceLabel(trimmed, { mode: 'compact' })

  const rootClass = [
    'workspace-path',
    onOpenSelector
      ? 'workspace-path-open-selector'
      : expanded
        ? 'workspace-path-expanded'
        : 'workspace-path-collapsed',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  const handlePrimary = () => {
    if (onOpenSelector) {
      onOpenSelector()
      return
    }
    toggle()
  }

  return (
    <div className={rootClass} data-testid="workspace" title={trimmed}>
      <button
        type="button"
        className="workspace-path-toggle"
        data-testid={onOpenSelector ? 'workspace-open-selector' : 'workspace-toggle'}
        aria-expanded={onOpenSelector ? undefined : expanded}
        aria-label={
          onOpenSelector
            ? 'Open workspace path selector'
            : expanded
              ? 'Collapse workspace path'
              : 'Expand workspace path'
        }
        onClick={handlePrimary}
      >
        <span className="workspace-path-label" data-testid="workspace-label">
          {label}
        </span>
        {onOpenSelector ? (
          <span className="workspace-path-chevron" aria-hidden="true">
            ›
          </span>
        ) : null}
      </button>
      {copyVisible ? (
        <button
          type="button"
          className="workspace-path-copy"
          data-testid="workspace-copy"
          aria-label="Copy workspace path"
          title="Copy full path"
          onClick={(e) => {
            e.stopPropagation()
            void copyPath()
          }}
        >
          {copied ? 'Copied' : 'Copy'}
        </button>
      ) : null}
    </div>
  )
}
