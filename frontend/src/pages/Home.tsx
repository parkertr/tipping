import { Container, Typography, Paper, Grid, Box, Chip, CircularProgress, Alert } from '@mui/material'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import GoogleSignIn from '../components/GoogleSignIn'
import api from '../utils/auth'

interface UpcomingMatch {
  id: string
  homeTeam: string
  awayTeam: string
  date: string
  competition: string
  status: string
}

const Home = () => {
  const navigate = useNavigate()
  const { user, isAuthenticated } = useAuth()

  const { data: upcomingMatches, isLoading, error } = useQuery<UpcomingMatch[]>({
    queryKey: ['upcomingMatches'],
    queryFn: async () => {
      const response = await api.get('/matches/upcoming')
      return response.data
    },
  })

  const formatMatchDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffTime = date.getTime() - now.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))

    if (diffDays === 0) {
      return `Today at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    } else if (diffDays === 1) {
      return `Tomorrow at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    } else if (diffDays < 7) {
      return `${date.toLocaleDateString([], { weekday: 'long' })} at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    } else {
      return date.toLocaleDateString([], {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    }
  }

  // Show login prompt for unauthenticated users
  if (!isAuthenticated) {
    return (
      <Container maxWidth="sm">
        <Paper sx={{ p: 4, textAlign: 'center', mt: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Welcome to Footy Tipping
          </Typography>
          <Typography variant="body1" color="text.secondary" paragraph>
            Join our football tipping competition! Make predictions on upcoming matches,
            compete with friends, and climb the leaderboard.
          </Typography>
          <Box sx={{ mt: 3 }}>
            <GoogleSignIn />
          </Box>
        </Paper>
      </Container>
    )
  }

  return (
    <Container maxWidth="lg">
      <Typography variant="h4" component="h1" gutterBottom>
        Welcome back, {user?.name?.split(' ')[0] || 'Champion'}!
      </Typography>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="h5">
                Upcoming Matches
              </Typography>
              <Typography
                variant="body2"
                color="primary"
                sx={{ cursor: 'pointer', textDecoration: 'underline' }}
                onClick={() => navigate('/matches')}
              >
                View All Matches
              </Typography>
            </Box>

            {isLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
                <CircularProgress />
              </Box>
            ) : error ? (
              <Alert severity="error" sx={{ mb: 2 }}>
                Failed to load upcoming matches. Please try again later.
              </Alert>
            ) : upcomingMatches && upcomingMatches.length > 0 ? (
              upcomingMatches.map((match) => (
                <Paper
                  key={match.id}
                  sx={{
                    p: 2,
                    mb: 2,
                    cursor: 'pointer',
                    transition: 'all 0.2s ease-in-out',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                      boxShadow: 3,
                    }
                  }}
                  onClick={() => navigate('/matches')}
                >
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1 }}>
                    <Typography variant="h6" sx={{ fontWeight: 600 }}>
                      {match.homeTeam} vs {match.awayTeam}
                    </Typography>
                    <Chip
                      label={match.competition}
                      size="small"
                      color="primary"
                      variant="outlined"
                    />
                  </Box>
                  <Typography color="text.secondary" variant="body2">
                    {formatMatchDate(match.date)}
                  </Typography>
                </Paper>
              ))
            ) : (
              <Box sx={{ textAlign: 'center', py: 4 }}>
                <Typography color="text.secondary">
                  No upcoming matches scheduled
                </Typography>
              </Box>
            )}
          </Paper>
        </Grid>

        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              Quick Stats
            </Typography>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography color="text.secondary">Current Position:</Typography>
                <Chip label="5th" color="secondary" size="small" />
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography color="text.secondary">Points This Week:</Typography>
                <Typography variant="h6" color="primary">8</Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography color="text.secondary">Total Points:</Typography>
                <Typography variant="h6" color="primary">45</Typography>
              </Box>
            </Box>
          </Paper>

          <Paper sx={{ p: 3, mt: 2 }}>
            <Typography variant="h6" gutterBottom>
              Quick Actions
            </Typography>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
              <Typography
                variant="body2"
                color="primary"
                sx={{ cursor: 'pointer', textDecoration: 'underline' }}
                onClick={() => navigate('/matches')}
              >
                Make Predictions
              </Typography>
              <Typography
                variant="body2"
                color="primary"
                sx={{ cursor: 'pointer', textDecoration: 'underline' }}
                onClick={() => navigate('/leaderboard')}
              >
                View Leaderboard
              </Typography>
              <Typography
                variant="body2"
                color="primary"
                sx={{ cursor: 'pointer', textDecoration: 'underline' }}
                onClick={() => navigate('/profile')}
              >
                View Profile
              </Typography>
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  )
}

export default Home
