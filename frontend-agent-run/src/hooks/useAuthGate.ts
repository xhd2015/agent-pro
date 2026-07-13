import { useCallback, useEffect, useState } from 'react'
import { clearToken, fetchHealth, getToken } from '../api/client'

export function useAuthGate() {
  const [needsAuth, setNeedsAuth] = useState(() => !getToken())
  const [ready, setReady] = useState(() => Boolean(getToken()))

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      const token = getToken()
      const status = await fetchHealth()
      if (cancelled) return
      if (status === 401) {
        if (token) clearToken()
        setNeedsAuth(true)
        setReady(true)
        return
      }
      setNeedsAuth(false)
      setReady(true)
    })()
    return () => {
      cancelled = true
    }
  }, [])

  const markAuthenticated = useCallback(() => {
    setNeedsAuth(false)
    setReady(true)
  }, [])

  return { needsAuth, ready, markAuthenticated }
}
