import { useState } from 'react'
import { Container, Typography, Paper, Grid, TextField, Button, List, ListItem, ListItemText } from '@mui/material'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'

interface UserProfile {
  username: string
  email: string
  joinDate: string
  stats: {
    totalPoints: number
    correctPredictions: number
    totalPredictions: number
    currentRank: number
  }
  recentPredictions: {
    matchId: string
    homeTeam: string
    awayTeam: string
    prediction: string
    result: string
    points: number
  }[]
}

const Profile = () => {
  const [isEditing, setIsEditing] = useState(false)
  const [email, setEmail] = useState('')
  const queryClient = useQueryClient()

  const { data: profile, isLoading } = useQuery<UserProfile>({
    queryKey: ['profile'],
    queryFn: async () => {
      const response = await axios.get('/api/profile')
      setEmail(response.data.email)
      return response.data
    },
  })

  const updateProfile = useMutation({
    mutationFn: async (newEmail: string) => {
      const response = await axios.put('/api/profile', { email: newEmail })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profile'] })
      setIsEditing(false)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateProfile.mutate(email)
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
            <Typography>Username: {profile?.username}</Typography>
            <Typography>Join Date: {new Date(profile?.joinDate || '').toLocaleDateString()}</Typography>

            {isEditing ? (
              <form onSubmit={handleSubmit}>
                <TextField
                  fullWidth
                  label="Email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  margin="normal"
                />
                <Button type="submit" variant="contained" sx={{ mr: 1 }}>
                  Save
                </Button>
                <Button variant="outlined" onClick={() => setIsEditing(false)}>
                  Cancel
                </Button>
              </form>
            ) : (
              <>
                <Typography>Email: {profile?.email}</Typography>
                <Button variant="outlined" onClick={() => setIsEditing(true)} sx={{ mt: 2 }}>
                  Edit Profile
                </Button>
              </>
            )}
          </Paper>
        </Grid>

        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              Statistics
            </Typography>
            <Typography>Total Points: {profile?.stats.totalPoints}</Typography>
            <Typography>Correct Predictions: {profile?.stats.correctPredictions}</Typography>
            <Typography>Success Rate: {((profile?.stats.correctPredictions || 0) / (profile?.stats.totalPredictions || 1) * 100).toFixed(1)}%</Typography>
            <Typography>Current Rank: {profile?.stats.currentRank}</Typography>
          </Paper>
        </Grid>

        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              Recent Predictions
            </Typography>
            <List>
              {profile?.recentPredictions.map((prediction) => (
                <ListItem key={prediction.matchId}>
                  <ListItemText
                    primary={`${prediction.homeTeam} vs ${prediction.awayTeam}`}
                    secondary={`Prediction: ${prediction.prediction} | Result: ${prediction.result} | Points: ${prediction.points}`}
                  />
                </ListItem>
              ))}
            </List>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  )
}

export default Profile
