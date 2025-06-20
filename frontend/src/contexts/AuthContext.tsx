import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { googleLogout, CredentialResponse } from '@react-oauth/google'
import Cookies from 'js-cookie'

interface User {
  id: string
  googleId: string
  email: string
  name: string
  pictureUrl: string
  isActive: boolean
}

interface AuthContextType {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (credentialResponse: CredentialResponse) => Promise<void>
  logout: () => void
  refreshToken: () => Promise<void>
  updateProfile: (updates: Partial<User>) => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

interface AuthProviderProps {
  children: ReactNode
}

const API_BASE_URL = 'http://localhost:8080/api'

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const isAuthenticated = !!user

  // Check for existing token on app start
  useEffect(() => {
    const token = Cookies.get('auth_token')
    if (token) {
      // Verify token and get user profile
      fetchCurrentUser()
    } else {
      setIsLoading(false)
    }
  }, [])

  const fetchCurrentUser = async () => {
    try {
      const token = Cookies.get('auth_token')
      if (!token) {
        setIsLoading(false)
        return
      }

      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      })

      if (response.ok) {
        const userData = await response.json()
        setUser(userData)
      } else {
        // Token is invalid, remove it
        Cookies.remove('auth_token')
        setUser(null)
      }
    } catch (error) {
      console.error('Error fetching current user:', error)
      Cookies.remove('auth_token')
      setUser(null)
    } finally {
      setIsLoading(false)
    }
  }

    const login = async (credentialResponse: CredentialResponse) => {
    try {
      setIsLoading(true)

      // Send the Google credential to our backend
      const response = await fetch(`${API_BASE_URL}/auth/google/callback`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          credential: credentialResponse.credential,
        }),
      })

      if (response.ok) {
        const { user: userData, token } = await response.json()

        // Store the JWT token in a secure cookie
        Cookies.set('auth_token', token, {
          expires: 7, // 7 days
          secure: false, // Set to true in production
          sameSite: 'strict'
        })

        setUser(userData)
      } else {
        throw new Error('Authentication failed')
      }
    } catch (error) {
      console.error('Login error:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const logout = () => {
    // Remove token from cookies
    Cookies.remove('auth_token')

    // Clear user state
    setUser(null)

    // Logout from Google
    googleLogout()
  }

  const refreshToken = async () => {
    try {
      const currentToken = Cookies.get('auth_token')
      if (!currentToken) {
        throw new Error('No token to refresh')
      }

      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${currentToken}`,
          'Content-Type': 'application/json',
        },
      })

      if (response.ok) {
        const { token } = await response.json()

        Cookies.set('auth_token', token, {
          expires: 7,
          secure: false, // Set to true in production
          sameSite: 'strict'
        })
      } else {
        // Refresh failed, logout user
        logout()
        throw new Error('Token refresh failed')
      }
    } catch (error) {
      console.error('Token refresh error:', error)
      logout()
      throw error
    }
  }

  const updateProfile = async (updates: Partial<User>) => {
    try {
      const token = Cookies.get('auth_token')
      if (!token) {
        throw new Error('No authentication token')
      }

      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      })

      if (response.ok) {
        const updatedUser = await response.json()
        setUser(updatedUser)
      } else {
        throw new Error('Profile update failed')
      }
    } catch (error) {
      console.error('Profile update error:', error)
      throw error
    }
  }

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated,
    login,
    logout,
    refreshToken,
    updateProfile,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
