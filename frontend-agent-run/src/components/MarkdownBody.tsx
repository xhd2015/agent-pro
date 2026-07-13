import type { AnchorHTMLAttributes } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import './MarkdownBody.css'

export type MarkdownBodyProps = {
  text: string
  className?: string
  /** Compact styling for progress/thinking cards */
  compact?: boolean
}

function externalLinkProps(href?: string): AnchorHTMLAttributes<HTMLAnchorElement> {
  if (!href) return {}
  // Block javascript: and other non-http(s)/relative schemes for safety.
  const trimmed = href.trim()
  const lower = trimmed.toLowerCase()
  if (
    lower.startsWith('javascript:') ||
    lower.startsWith('data:') ||
    lower.startsWith('vbscript:')
  ) {
    return { href: undefined }
  }
  const isExternal = /^https?:\/\//i.test(trimmed)
  if (isExternal) {
    return { target: '_blank', rel: 'noopener noreferrer' }
  }
  return {}
}

/**
 * Renders markdown for assistant responses and thinking cards.
 * Uses react-markdown (no raw HTML by default) + GFM (tables, strikethrough, etc.).
 */
export function MarkdownBody({ text, className = '', compact = false }: MarkdownBodyProps) {
  const content = text ?? ''
  const classes = ['markdown-body', compact ? 'markdown-body-compact' : '', className]
    .filter(Boolean)
    .join(' ')

  return (
    <div className={classes} data-testid="markdown-body">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          a: ({ href, children, ...rest }) => (
            <a href={href} {...externalLinkProps(href)} {...rest}>
              {children}
            </a>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
