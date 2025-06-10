import React, { useState, ChangeEvent } from 'react'
import { Container, Typography, Paper, Grid, Button, TextField } from '@mui/material'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'

interface Match {
  id: string
  homeTeam: string
  awayTeam: string
  date: string
  competition: string
  prediction?: string
}

const Matches: React.FC = () => {
  const [selectedMatch, setSelectedMatch] = useState<string | null>(null)
  const [prediction, setPrediction] = useState('')
  const queryClient = useQueryClient()

  const { data: matches, isLoading } = useQuery<Match[]>({
    queryKey: ['matches'],
    queryFn: async () => {
      const response = await axios.get('/api/matches')
      return response.data
    },
  })

  const submitPrediction = useMutation({
    mutationFn: async ({ matchId, prediction }: { matchId: string; prediction: string }) => {
      const response = await axios.post(`/api/matches/${matchId}/predict`, { prediction })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['matches'] })
      setSelectedMatch(null)
      setPrediction('')
    },
  })

  const handleSubmit = (matchId: string) => {
    submitPrediction.mutate({ matchId, prediction })
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
          matches?.map((match: Match) => (
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
                          />
                        </Grid>
                        <Grid item xs={4}>
                          <Button
                            variant="contained"
                            onClick={() => handleSubmit(match.id)}
                            disabled={!prediction}
                          >
                            Submit
                          </Button>
                        </Grid>
                      </Grid>
                    ) : (
                      <Button
                        variant="outlined"
                        onClick={() => setSelectedMatch(match.id)}
                        disabled={!!match.prediction}
                      >
                        {match.prediction ? `Predicted: ${match.prediction}` : 'Make Prediction'}
                      </Button>
                    )}
                  </Grid>
                </Grid>
              </Paper>
            </Grid>
          ))
        )}
      </Grid>
    </Container>
  )
}

export default Matches
