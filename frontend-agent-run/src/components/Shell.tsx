import type { ReactNode } from 'react'
import './Shell.css'

export function Shell({
  children,
  sessionPage = false,
  homePage = false,
}: {
  children: ReactNode
  sessionPage?: boolean
  homePage?: boolean
}) {
  const pageClass = sessionPage ? ' session-page' : homePage ? ' home-page' : ''
  return (
    <div className={`app-shell${pageClass}`} data-testid="app-shell">
      {children}
    </div>
  )
}
