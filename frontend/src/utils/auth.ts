import axios from 'axios'
import Cookies from 'js-cookie'

const API_BASE_URL = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`

// Create axios instance with base configuration
export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Token management functions
export const getAuthToken = (): string | null => {
  return Cookies.get('auth_token') || null
}

export const setAuthToken = (token: string): void => {
  Cookies.set('auth_token', token, {
    expires: 7, // 7 days
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict'
  })
}

export const removeAuthToken = (): void => {
  Cookies.remove('auth_token')
}

// Axios request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = getAuthToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Axios response interceptor for handling 401 errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      // Try to refresh token
      try {
        const currentToken = getAuthToken()
        if (currentToken) {
          const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${currentToken}`,
              'Content-Type': 'application/json',
            },
          })

          if (response.ok) {
            const { token } = await response.json()
            setAuthToken(token)

            // Retry the original request with new token
            originalRequest.headers.Authorization = `Bearer ${token}`
            return api(originalRequest)
          }
        }
      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError)
      }

      // If refresh fails, remove token and redirect to login
      removeAuthToken()
      window.location.href = '/'
    }

    return Promise.reject(error)
  }
)

export default api
