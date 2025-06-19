import { useState } from 'react'
import { Container, Typography, Paper, Grid, TextField, Button, List, ListItem, ListItemText, Avatar, Box } from '@mui/material'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '../contexts/AuthContext'
import api from '../utils/auth'

// Removed unused UserProfile interface

const Profile = () => {
  const [isEditing, setIsEditing] = useState(false)
  const [name, setName] = useState('')
  const { user, updateProfile: updateAuthProfile } = useAuth()
  const queryClient = useQueryClient()

  // Initialize form with user data
  useState(() => {
    if (user) {
      setName(user.name)
    }
  })

  const { data: userStats, isLoading } = useQuery({
    queryKey: ['userStats'],
    queryFn: async () => {
      const response = await api.get('/auth/me/stats')
      return response.data
    },
  })

  const updateProfile = useMutation({
    mutationFn: async (updates: { name: string }) => {
      await updateAuthProfile(updates)
      return updates
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['userStats'] })
      setIsEditing(false)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateProfile.mutate({ name })
  }

  if (isLoading) {
    return <Typography>Loading profile...</Typography>
  }

  return (
    <Container maxWidth="lg">
      <Typography variant="h4" component="h1" gutterBottom>
        Profile
      </Typography>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              User Information
            </Typography>

            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <Avatar
                src={user?.pictureUrl}
                alt={user?.name}
                sx={{ width: 64, height: 64, mr: 2 }}
              />
              <Box>
                <Typography variant="h6">{user?.name}</Typography>
                <Typography color="text.secondary">{user?.email}</Typography>
              </Box>
            </Box>

            {isEditing ? (
              <form onSubmit={handleSubmit}>
                <TextField
                  fullWidth
                  label="Name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  margin="normal"
                />
                <Box sx={{ mt: 2 }}>
                  <Button
                    type="submit"
                    variant="contained"
                    sx={{ mr: 1 }}
                    disabled={updateProfile.isPending}
                  >
                    {updateProfile.isPending ? 'Saving...' : 'Save'}
                  </Button>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      setIsEditing(false)
                      setName(user?.name || '')
                    }}
                  >
                    Cancel
                  </Button>
                </Box>
              </form>
            ) : (
              <Button variant="outlined" onClick={() => setIsEditing(true)} sx={{ mt: 2 }}>
                Edit Profile
              </Button>
            )}
          </Paper>
        </Grid>

        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              Statistics
            </Typography>
            {isLoading ? (
              <Typography>Loading stats...</Typography>
            ) : (
              <>
                <Typography>Total Points: {userStats?.totalPoints || 0}</Typography>
                <Typography>Correct Predictions: {userStats?.correctPredictions || 0}</Typography>
                <Typography>
                  Success Rate: {
                    userStats?.totalPredictions > 0
                      ? ((userStats.correctPredictions / userStats.totalPredictions) * 100).toFixed(1)
                      : 0
                  }%
                </Typography>
                <Typography>Current Rank: {userStats?.currentRank || 'Unranked'}</Typography>
              </>
            )}
          </Paper>
        </Grid>

        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              Recent Predictions
            </Typography>
            {isLoading ? (
              <Typography>Loading predictions...</Typography>
            ) : userStats?.recentPredictions?.length > 0 ? (
              <List>
                {userStats.recentPredictions.map((prediction: any, index: number) => (
                  <ListItem key={`${prediction.matchId}-${index}`}>
                    <ListItemText
                      primary={`${prediction.homeTeam} vs ${prediction.awayTeam}`}
                      secondary={`Prediction: ${prediction.prediction} | Result: ${prediction.result || 'Pending'} | Points: ${prediction.points || 0}`}
                    />
                  </ListItem>
                ))}
              </List>
            ) : (
              <Typography color="text.secondary">
                No predictions yet. Start making predictions on upcoming matches!
              </Typography>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Container>
  )
}

export default Profile
