import { Container, Typography, Paper, Grid } from '@mui/material'
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'

interface UpcomingMatch {
  id: string
  homeTeam: string
  awayTeam: string
  date: string
  competition: string
}

const Home = () => {
  const { data: upcomingMatches, isLoading } = useQuery<UpcomingMatch[]>({
    queryKey: ['upcomingMatches'],
    queryFn: async () => {
      const response = await axios.get('/api/matches/upcoming')
      return response.data
    },
  })

  return (
    <Container maxWidth="lg">
      <Typography variant="h4" component="h1" gutterBottom>
        Welcome to Footy Tipping
      </Typography>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              Upcoming Matches
            </Typography>
            {isLoading ? (
              <Typography>Loading matches...</Typography>
            ) : (
              upcomingMatches?.map((match) => (
                <Paper key={match.id} sx={{ p: 2, mb: 2 }}>
                  <Typography variant="h6">
                    {match.homeTeam} vs {match.awayTeam}
                  </Typography>
                  <Typography color="text.secondary">
                    {new Date(match.date).toLocaleDateString()} - {match.competition}
                  </Typography>
                </Paper>
              ))
            )}
          </Paper>
        </Grid>

        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              Quick Stats
            </Typography>
            <Typography>Your current position: 5th</Typography>
            <Typography>Points this week: 8</Typography>
            <Typography>Total points: 45</Typography>
          </Paper>
        </Grid>
      </Grid>
    </Container>
  )
}

export default Home
