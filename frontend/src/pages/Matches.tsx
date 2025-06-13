import React, { useState, ChangeEvent, useEffect } from 'react'
import { Container, Typography, Paper, Grid, Button, TextField } from '@mui/material'
import { useQuery, useMutation } from '@tanstack/react-query'
import axios from 'axios'

interface Match {
  id: string
  homeTeam: string
  awayTeam: string
  date: string
  competition: string
}

interface Prediction {
  id: string
  userId: string
  matchId: string
  homeGoals: number
  awayGoals: number
  createdAt: string
  points: number
}

const Matches: React.FC = () => {
  const [selectedMatch, setSelectedMatch] = useState<string | null>(null)
  const [prediction, setPrediction] = useState('')
  const [userPredictions, setUserPredictions] = useState<Record<string, Prediction>>({})

  // TODO: Replace with actual user ID from authentication
  const currentUserId = 'user123'

  const { data: matches, isLoading } = useQuery<Match[]>({
    queryKey: ['matches'],
    queryFn: async () => {
      const response = await axios.get('/api/matches')
      return response.data
    },
  })

  // Fetch user predictions for all matches
  useEffect(() => {
    const fetchPredictions = async () => {
      if (!matches) return

      const predictions: Record<string, Prediction> = {}

      for (const match of matches) {
        try {
          const response = await axios.get(`/api/matches/${match.id}/predictions/${currentUserId}`)
          predictions[match.id] = response.data
        } catch (error) {
          // No prediction found for this match, which is fine
        }
      }

      setUserPredictions(predictions)
    }

    fetchPredictions()
  }, [matches, currentUserId])

  const submitPrediction = useMutation({
    mutationFn: async ({ matchId, prediction }: { matchId: string; prediction: string }) => {
      // Parse prediction format "2-1" into homeGoals and awayGoals
      const scoreParts = prediction.split('-')
      if (scoreParts.length !== 2) {
        throw new Error('Invalid prediction format. Please use format like "2-1"')
      }

      const homeGoals = parseInt(scoreParts[0].trim())
      const awayGoals = parseInt(scoreParts[1].trim())

      if (isNaN(homeGoals) || isNaN(awayGoals)) {
        throw new Error('Invalid prediction format. Please use numbers like "2-1"')
      }

      // Use the correct API endpoint and format
      const response = await axios.post('/api/predictions', {
        userId: currentUserId,
        matchId,
        homeGoals,
        awayGoals
      })
      return response.data
    },
    onSuccess: (data, variables) => {
      // Update local predictions state
      setUserPredictions(prev => ({
        ...prev,
        [variables.matchId]: data
      }))
      setSelectedMatch(null)
      setPrediction('')
    },
    onError: (error: any) => {
      console.error('Failed to submit prediction:', error)
      alert(error.message || 'Failed to submit prediction')
    },
  })

  const handleSubmit = (matchId: string) => {
    submitPrediction.mutate({ matchId, prediction })
  }

  const formatPrediction = (pred: Prediction) => {
    return `${pred.homeGoals}-${pred.awayGoals}`
  }

  return (
    <Container maxWidth="lg">
      <Typography variant="h4" component="h1" gutterBottom>
        Matches
      </Typography>

      <Grid container spacing={3}>
        {isLoading ? (
          <Typography>Loading matches...</Typography>
        ) : (
          matches?.map((match: Match) => {
            const existingPrediction = userPredictions[match.id]

            return (
              <Grid item xs={12} key={match.id}>
                <Paper sx={{ p: 3 }}>
                  <Grid container spacing={2} alignItems="center">
                    <Grid item xs={12} md={6}>
                      <Typography variant="h6">
                        {match.homeTeam} vs {match.awayTeam}
                      </Typography>
                      <Typography color="text.secondary">
                        {new Date(match.date).toLocaleDateString()} - {match.competition}
                      </Typography>
                    </Grid>
                    <Grid item xs={12} md={6}>
                      {selectedMatch === match.id ? (
                        <Grid container spacing={2}>
                          <Grid item xs={8}>
                            <TextField
                              fullWidth
                              label="Your Prediction"
                              value={prediction}
                              onChange={(e: ChangeEvent<HTMLInputElement>) => setPrediction(e.target.value)}
                              placeholder="e.g., 2-1"
                              helperText="Enter score prediction (e.g., 2-1)"
                            />
                          </Grid>
                          <Grid item xs={4}>
                            <Button
                              variant="contained"
                              onClick={() => handleSubmit(match.id)}
                              disabled={!prediction || submitPrediction.isPending}
                            >
                              {submitPrediction.isPending ? 'Submitting...' : 'Submit'}
                            </Button>
                          </Grid>
                        </Grid>
                      ) : (
                        <Button
                          variant="outlined"
                          onClick={() => setSelectedMatch(match.id)}
                          disabled={!!existingPrediction}
                        >
                          {existingPrediction
                            ? `Predicted: ${formatPrediction(existingPrediction)}`
                            : 'Make Prediction'
                          }
                        </Button>
                      )}
                    </Grid>
                  </Grid>
                </Paper>
              </Grid>
            )
          })
        )}
      </Grid>
    </Container>
  )
}

export default Matches
