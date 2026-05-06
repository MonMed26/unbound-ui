import { useState, useEffect, useCallback } from 'react'
import { authApi } from '../api/client'

interface AuthState {
  isAuthenticated: boolean
  isLoading: boolean
  setupRequired: boolean
}

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    isAuthenticated: !!localStorage.getItem('token'),
    isLoading: true,
    setupRequired: false,
  })

  useEffect(() => {
    checkAuth()
  }, [])

  const checkAuth = async () => {
    try {
      const { data } = await authApi.status()
      setState({
        isAuthenticated: !!localStorage.getItem('token'),
        isLoading: false,
        setupRequired: data.setup_required,
      })
    } catch {
      setState({
        isAuthenticated: false,
        isLoading: false,
        setupRequired: false,
      })
    }
  }

  const login = useCallback(async (username: string, password: string) => {
    const { data } = await authApi.login(username, password)
    localStorage.setItem('token', data.token)
    setState((prev) => ({ ...prev, isAuthenticated: true }))
  }, [])

  const setup = useCallback(async (username: string, password: string) => {
    const { data } = await authApi.setup(username, password)
    localStorage.setItem('token', data.token)
    setState({ isAuthenticated: true, isLoading: false, setupRequired: false })
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    setState((prev) => ({ ...prev, isAuthenticated: false }))
  }, [])

  return { ...state, login, setup, logout }
}
