import { useState } from 'react'
import { Route, Routes, useLocation } from 'react-router-dom'
import './App.css'
import { Shell } from './components/Shell'
import {
  initialHomeSessionListCache,
  type HomeSessionListCache,
} from './homeSessionListCache'
import { useAuthGate } from './hooks/useAuthGate'
import { AuthPage } from './pages/AuthPage'
import { HomePage } from './pages/HomePage'
import { NotFoundPage } from './pages/NotFoundPage'
import { SessionPage } from './pages/SessionPage'
import { WorkspacePage } from './pages/WorkspacePage'

/**
 * Keep Home mounted while visiting session/workspace so the list DOM scroll
 * position is preserved. List data also lives at App level.
 *
 * Hide inactive home with opacity/pointer-events only (not display:none /
 * visibility:hidden), which can reset overflow scrollTop in some engines.
 */
function AppRoutes() {
  const location = useLocation()
  const [homeDraft, setHomeDraft] = useState('')
  const [homeListCache, setHomeListCache] = useState<HomeSessionListCache>(
    initialHomeSessionListCache,
  )

  const path = location.pathname
  const isHome = path === '/'
  const isWorkspace = path === '/workspace'
  const isSession = path.startsWith('/sessions/')
  const showOverlay = isWorkspace || isSession || (!isHome && !isWorkspace && !isSession)

  return (
    <div className="app-routes-root" data-testid="app-routes-root">
      <div
        className="app-home-keepalive"
        data-testid="app-home-keepalive"
        data-active={isHome ? 'true' : 'false'}
        aria-hidden={isHome ? undefined : true}
        style={{
          position: isHome ? 'relative' : 'absolute',
          inset: isHome ? undefined : 0,
          display: 'flex',
          flexDirection: 'column',
          flex: isHome ? 1 : undefined,
          minHeight: 0,
          height: '100%',
          width: '100%',
          opacity: isHome ? 1 : 0,
          pointerEvents: isHome ? 'auto' : 'none',
          zIndex: isHome ? 1 : 0,
        }}
      >
        <HomePage
          draft={homeDraft}
          onDraftChange={setHomeDraft}
          listCache={homeListCache}
          onListCacheChange={setHomeListCache}
          isActive={isHome}
        />
      </div>

      {showOverlay ? (
        <div
          className="app-overlay-route"
          data-testid="app-overlay-route"
          style={{
            position: 'absolute',
            inset: 0,
            zIndex: 2,
            display: 'flex',
            flexDirection: 'column',
            minHeight: 0,
            height: '100%',
          }}
        >
          <Routes>
            <Route path="/workspace" element={<WorkspacePage />} />
            <Route path="/sessions/:sessionId" element={<SessionPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </div>
      ) : null}
    </div>
  )
}

export default function App() {
  const { needsAuth, ready, markAuthenticated } = useAuthGate()

  // Auth gate (not a URL). Soft flip after token submit — no hard reload.
  if (needsAuth) {
    return <AuthPage onAuthenticated={markAuthenticated} />
  }

  if (!ready) {
    return (
      <Shell>
        <div className="main-panel" />
      </Shell>
    )
  }

  return <AppRoutes />
}
