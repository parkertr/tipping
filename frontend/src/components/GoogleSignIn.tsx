import React from 'react'
import { GoogleLogin } from '@react-oauth/google'
import { Box, Alert } from '@mui/material'
import { useAuth } from '../contexts/AuthContext'

interface GoogleSignInProps {
  onError?: (error: string) => void
}

const GoogleSignIn: React.FC<GoogleSignInProps> = ({ onError }) => {
  const { login } = useAuth()
  const [error, setError] = React.useState<string | null>(null)

  const handleSuccess = async (credentialResponse: any) => {
    try {
      setError(null)
      await login(credentialResponse)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Login failed'
      setError(errorMessage)
      onError?.(errorMessage)
    }
  }

  const handleError = () => {
    const errorMessage = 'Google Sign-In failed. Please try again.'
    setError(errorMessage)
    onError?.(errorMessage)
  }

  return (
    <Box>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      <GoogleLogin
        onSuccess={handleSuccess}
        onError={handleError}
        theme="outline"
        size="large"
        text="signin_with"
        shape="rectangular"
      />
    </Box>
  )
}

export default GoogleSignIn
